package controllers

import (
    "context"
    "fmt"

    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/log"
    examplev1 "github.com/your-username/k8s-operator-example/api/v1"
)

// ImageArrayReconciler reconciles an ImageArray object
type ImageArrayReconciler struct {
    client.Client
}

// Reconcile method for the ImageArray
func (r *ImageArrayReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    logger := log.FromContext(ctx)

    // Fetch the ImageArray instance
    imageArray := &examplev1.ImageArray{}
    if err := r.Get(ctx, req.NamespacedName, imageArray); err != nil {
        logger.Error(err, "Failed to get ImageArray")
        return reconcile.Result{}, client.IgnoreNotFound(err)
    }

    // Define the job name
    jobName := fmt.Sprintf("handle-images-%s", imageArray.Name)

    // Check if the job already exists
    foundJob := &batchv1.Job{}
    err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: jobName}, foundJob)
    if err == nil {
        // Job already exists
        logger.Info("Job already exists", "Job.Namespace", foundJob.Namespace, "Job.Name", foundJob.Name)
        return reconcile.Result{}, nil
    }

    // Create a new Job
    job := r.createImageHandlingJob(imageArray, jobName, req.Namespace)

    // Set the ImageArray as the owner of the Job
    if err := controllerutil.SetControllerReference(imageArray, job, r.Scheme()); err != nil {
        logger.Error(err, "Failed to set controller reference")
        return reconcile.Result{}, err
    }

    logger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
    if err := r.Create(ctx, job); err != nil {
        logger.Error(err, "Failed to create Job")
        return reconcile.Result{}, err
    }

    return reconcile.Result{}, nil
}

// createImageHandlingJob creates a Kubernetes job that pulls, tags, and pushes images
func (r *ImageArrayReconciler) createImageHandlingJob(imageArray *examplev1.ImageArray, jobName, namespace string) *batchv1.Job {
    // Generate the shell commands to pull, tag, and push images
    var commands []string
    for _, image := range imageArray.Spec.Images {
        sourceImage := fmt.Sprintf("%s/%s", "registry-1", image)
        targetImage := fmt.Sprintf("%s/%s", "registry-2", image)

        commands = append(commands,
            fmt.Sprintf("docker pull %s", sourceImage),
            fmt.Sprintf("docker tag %s %s", sourceImage, targetImage),
            fmt.Sprintf("docker push %s", targetImage),
        )
    }
    // Add docker login commands from secrets
    commands = append([]string{
        "docker login registry-1 --username=$(cat /etc/secret/source-registry-username) --password=$(cat /etc/secret/source-registry-password)",
        "docker login registry-2 --username=$(cat /etc/secret/target-registry-username) --password=$(cat /etc/secret/target-registry-password)",
    }, commands...)

    // Join commands into a single script
    script := fmt.Sprintf("set -e; %s", commands)

    // Create the job
    return &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      jobName,
            Namespace: namespace,
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:    "image-handler",
                            Image:   "docker:latest",  // Use Docker-in-Docker image
                            Command: []string{"sh", "-c", script},
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      "source-secret",
                                    MountPath: "/etc/secret/source",
                                },
                                {
                                    Name:      "target-secret",
                                    MountPath: "/etc/secret/target",
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume{
                        {
                            Name: "source-secret",
                            VolumeSource: corev1.VolumeSource{
                                Secret: &corev1.SecretVolumeSource{
                                    SecretName: "source-registry-image-pull-secret",
                                },
                            },
                        },
                        {
                            Name: "target-secret",
                            VolumeSource: corev1.VolumeSource{
                                Secret: &corev1.SecretVolumeSource{
                                    SecretName: "target-registry-image-push-secret",
                                },
                            },
                        },
                    },
                    RestartPolicy: corev1.RestartPolicyOnFailure,
                },
            },
        },
    }
}
