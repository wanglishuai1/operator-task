package builder

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/shenyisyn/myci/pkg/apis/task/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

type PodBuilder struct {
	task *v1alpha1.Task // 任务对象
	client.Client
}

// 构造函数
func NewPodBuilder(task *v1alpha1.Task, client client.Client) *PodBuilder {
	return &PodBuilder{task: task, Client: client}
}

// 本课程来自 程序员在囧途(www.jtthink.com) 咨询群：98514334
func (pb *PodBuilder) setInitContainer(pod *v1.Pod) {
	pod.Spec.InitContainers = []v1.Container{
		{
			Name:            pod.Name + "init",
			Image:           "shenyisyn/entrypoint:v1",
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         []string{"cp", "/app/entrypoint", "/entrypoint/bin"},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "entrypoint-volume",
					MountPath: "/entrypoint/bin",
				},
			},
		},
	}
}

// 本课程来自 程序员在囧途(www.jtthink.com) 咨询群：98514334
func (pb *PodBuilder) setPodVolumes(pod *v1.Pod) {
	volumes := []v1.Volume{
		{
			Name: "entrypoint-volume",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "podinfo",
			VolumeSource: v1.VolumeSource{
				DownwardAPI: &v1.DownwardAPIVolumeSource{
					Items: []v1.DownwardAPIVolumeFile{
						{
							Path: "order",
							FieldRef: &v1.ObjectFieldSelector{
								FieldPath: "metadata.annotations['taskorder']",
							},
						},
					},
				},
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volumes...)
}

const (
	//入口镜像  harbor
	EntryPointImage              = "docker.io/shenyisyn/entrypoint:v1.1"
	TaskPodPrefix                = "task-pod-" //Task对应的POD名称前缀
	AnnotationTaskOrderKey       = "taskorder" //要创建的 注解名称
	AnnotationTaskOrderInitValue = "0"
	AnnotationExitOrder          = "-1" //退出step用的Order标识
)

// 设置 POD元信息 包含 注解
func (pb *PodBuilder) setPodMeta(pod *v1.Pod) {
	pod.Namespace = pb.task.Namespace
	pod.Name = TaskPodPrefix + pb.task.Name        // pod名称
	pod.Spec.RestartPolicy = v1.RestartPolicyNever //从不 重启
	pod.Annotations = map[string]string{
		AnnotationTaskOrderKey: AnnotationTaskOrderInitValue,
	}
}

// 判断  Task对应的Pod是否存在， 如果存在 返回POD
func (pb *PodBuilder) getChildPod() (*v1.Pod, error) {
	pod := &v1.Pod{}
	err := pb.Client.Get(context.Background(), types.NamespacedName{
		Namespace: pb.task.Namespace, Name: TaskPodPrefix + pb.task.Name,
	}, pod)
	if err != nil { //没取到POD，则要进入创建流程
		return nil, err
	}
	return pod, err

}

// 设置容器，注意，增加了error返回值，
func (pb *PodBuilder) setContainer(index int, step v1alpha1.TaskStep) (v1.Container, error) {
	// 这里要强烈注意：step.Command必须要设置，如果没设置则通过http 去远程取。取不到直接报错
	command := step.Command // 取出它 原始的command ,是个 string切片
	if len(command) == 0 {  //没有写 command  . 需要从网上去解析
		ref, err := name.ParseReference(step.Image, name.WeakValidation)
		if err != nil {
			return step.Container, err
		}
		//从缓存获取
		var getImage *Image
		if v, ok := ImageCache.Get(ref); ok { //代表 缓存是有的
			getImage = &v
		} else { //缓存没有的情况下
			img, err := ParseImage(step.Image) //解析镜像
			if err != nil {
				return step.Container, err
			}
			ImageCache.Add(img.Ref, *img) //加入缓存
			getImage = img
		}
		//懒得解析， 暂时先写死 OS=Linux/amd64
		tempOs := "linux/amd64"
		if imgObj, ok := getImage.Command[tempOs]; ok {
			command = imgObj.Command
			if len(step.Args) == 0 { // 覆盖args （假设有的话)
				step.Args = imgObj.Args
			}
		} else {
			return step.Container, fmt.Errorf("error image command")
		}

	}
	args := step.Args //取出它原始的 args

	step.Container.ImagePullPolicy = v1.PullIfNotPresent //强迫设置拉取策略
	step.Container.Command = []string{"/entrypoint/bin/entrypoint"}
	step.Container.Args = []string{
		"--wait", "/etc/podinfo/order",
		"--waitcontent", strconv.Itoa(index + 1),
		"--out", "stdout", // entrypoint 中 写上stdout 就会定向到标准输出
		"--command",
	}
	// "sh -c"
	step.Container.Args = append(step.Container.Args, strings.Join(command, " "))
	step.Container.Args = append(step.Container.Args, args...)
	//设置挂载点
	step.Container.VolumeMounts = []v1.VolumeMount{
		{
			Name:      "entrypoint-volume",
			MountPath: "/entrypoint/bin",
		},
		{
			Name:      "podinfo",
			MountPath: "/etc/podinfo",
		},
	}
	return step.Container, nil
}

func (pb *PodBuilder) forward(ctx context.Context, pod *v1.Pod) error {
	if pod.Status.Phase == v1.PodSucceeded {
		return nil
	}
	// Order值 ==-1  代表 有一个step出错了。不做处理。
	if pod.Annotations[AnnotationTaskOrderKey] == AnnotationExitOrder {
		return nil
	}
	order, err := strconv.Atoi(pod.Annotations[AnnotationTaskOrderKey])
	if err != nil {
		return err
	}
	// 长度相等 ，代表已经到了最后一个 。不需要前进
	if order == len(pod.Spec.Containers) {
		return nil
	}
	//代表 当前的容器可能在等待  或者正在运行
	containerState := pod.Status.ContainerStatuses[order-1].State
	if containerState.Terminated == nil {
		return nil
	} else {
		//代表非正常退出
		if containerState.Terminated.ExitCode != 0 {
			//吧Order 值改成 -1
			pod.Annotations[AnnotationTaskOrderKey] = AnnotationExitOrder
			return pb.Client.Update(ctx, pod)
			//pod.Status.Phase=v1.PodFailed
			//return pb.Client.Status().Update(ctx,pod)
		}
	}
	order++
	pod.Annotations[AnnotationTaskOrderKey] = strconv.Itoa(order)
	return pb.Client.Update(ctx, pod)
}

// 本课程来自 程序员在囧途(www.jtthink.com) 咨询群：98514334
// 构建 创建出 对应的POD
func (pb *PodBuilder) Build(ctx context.Context) error {
	//判断 POD是否存在
	getPod, err := pb.getChildPod()
	if err == nil { //代表POD已经被创建
		//这代表启动阶段
		if getPod.Status.Phase == v1.PodRunning {
			//起始状态
			if getPod.Annotations[AnnotationTaskOrderKey] == AnnotationTaskOrderInitValue {
				getPod.Annotations[AnnotationTaskOrderKey] = "1" //写死  。故意的
				return pb.Client.Update(ctx, getPod)
			} else {
				if err := pb.forward(ctx, getPod); err != nil {
					return err
				}
			}
		}
		fmt.Println("新状态", getPod.Status.Phase)
		fmt.Println("order是", getPod.Annotations[AnnotationTaskOrderKey])
		return nil
	}

	newPod := &v1.Pod{}

	pb.setPodMeta(newPod) //设置元信息 ，如name,namespace 和annotations(重要的一匹)

	pb.setInitContainer(newPod) // 设置 initContainers

	c := []v1.Container{} // 容器切片
	for index, step := range pb.task.Spec.Steps {
		getContainer, err := pb.setContainer(index, step)
		if err != nil {
			return err
		}
		c = append(c, getContainer) //修改容器的command和args ----后面还要改
	}
	newPod.Spec.Containers = c
	pb.setPodVolumes(newPod) // 设置pod数据卷--重要，包含了downwardAPI 和emptyDir

	//设置owner
	newPod.OwnerReferences = append(newPod.OwnerReferences,
		metav1.OwnerReference{
			APIVersion: pb.task.APIVersion,
			Kind:       pb.task.Kind,
			Name:       pb.task.Name,
			UID:        pb.task.UID,
		})
	return pb.Create(ctx, newPod)
}

// 本课程来自 程序员在囧途(www.jtthink.com) 咨询群：98514334
func (pb *PodBuilder) setStep(pod *v1.Pod) {
	pod.Annotations = map[string]string{
		"taskorder": "0",
	}

}

//本课程来自 程序员在囧途(www.jtthink.com) 咨询群：98514334
