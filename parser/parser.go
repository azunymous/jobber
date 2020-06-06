package parser

import (
	"fmt"
	"io"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	"strings"
)

func InjectJobs(w io.Writer, b []byte) {
	jobI, all := parseYaml(b)
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Strict: true,
	})

	for i, o := range all {
		fmt.Println("---")
		if _, isJob := jobI[i]; isJob {
			j := o.(*batch.Job)
			inject(j)
		}
		err := serializer.Encode(o, w)
		if err != nil {
			panic(err)
		}
	}
}

// parseYaml parses through a YAML file's Kubernetes resources returning an set of indexes of Job resources and all objects.
func parseYaml(fileR []byte) (map[int]struct{}, []runtime.Object) {
	fileAsString := string(fileR[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	jobIndexes := make(map[int]struct{}, 1)
	all := make([]runtime.Object, 0, len(sepYamlfiles))
	for i, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)

		if err != nil {
			println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
			continue
		}

		_, ok := obj.(*batch.Job)
		if ok {
			jobIndexes[i] = struct{}{}
		}

		all = append(all, obj)
	}
	return jobIndexes, all
}

func inject(j *batch.Job) {
	if len(j.Spec.Template.Spec.Containers) < 1 {
		return
	}
	jobName := j.ObjectMeta.Name
	sharedVolume := core.Volume{
		Name: "jobber",
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	j.Spec.Template.Spec.Volumes = append(j.Spec.Template.Spec.Volumes, sharedVolume)

	sharedVolumeMount := core.VolumeMount{
		Name:      "jobber",
		MountPath: "/jobber/",
	}
	j.Spec.Template.Spec.Volumes = append(j.Spec.Template.Spec.Volumes)

	firstContainer := &j.Spec.Template.Spec.Containers[0]
	firstContainer.VolumeMounts = append(firstContainer.VolumeMounts, sharedVolumeMount)

	jobber := core.Container{
		Name:  "jobber",
		Image: "azunymous/jobber",
		Args: []string{
			"monitor",
			"-n",
			firstContainer.Name,
			"-v",
			"1",
		},
		Env: []core.EnvVar{
			{
				Name:  "JOB_NAME",
				Value: jobName,
			},
			{
				Name: "POD_NAME",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "NAMESPACE_NAME",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
		VolumeMounts: []core.VolumeMount{sharedVolumeMount},
	}
	j.Spec.Template.Spec.Containers = append(j.Spec.Template.Spec.Containers, jobber)
}
