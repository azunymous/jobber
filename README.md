# Jobber

Monitors Kubernetes jobs, uploading output files before job completion.

## Purpose
When running tests as a Kubernetes job, often test reports or simulation
logs are useful. However, you cannot use `kubectl cp` commands on 
completed Job pods. 

Alternatives include: 

- Tar the files and print them to standard out
- Stall the job while you poll it in order to `kubectl cp` the files
- Wrap your job container with a script of sorts to upload any required files
somewhere accessible

Jobber takes the final option and turns it into a sidecar that requires no 
modification to your original container. This way you do not need to 
install any additional dependencies. It uploads any resources to any 
object storage compatible with Minio (most S3/object storage).

In exchange, your Kubernetes Job resource should use jobber as a sidecar
container with a shared empty dir directory; along with a service account
that can view pods in your namespace.

e.g
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: testjob
  labels:
    app: testapp
spec:
  backoffLimit: 0
  template:
    spec:
      containers:
        - name: testjob
          # Important job container goes here
          image: busybox
          command:
            - sh
          args:
            - -c
            - sleep 10 && echo "content" > /data/testfile.txt
      restartPolicy: Never

```
becomes
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: my-job
  labels:
    app: my-job
spec:
  backoffLimit: 0
  template:
    spec:
      containers:
        - name: testjob
          # Important job container goes here
          image: busybox
          command:
            - sh
          args:
            - -c
            - sleep 10 && echo "content" > /data/testfile.txt
          volumeMounts:
            - mountPath: /data
              name: data
        - name: jobber
          image: azunymous/jobber
          args:
            - monitor
            - --name
            - testjob
            - --upload-file
            - /data/testfile.txt
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NAMESPACE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: JOBBER_ENDPOINT
              value: "minio:9000"
            - name: JOBBER_ACCESS_KEY
              value: my-access-key
            - name: JOBBER_SECRET_KEY
              value: should-be-from-a-secret
          volumeMounts:
            - mountPath: /data
              name: data
      volumes:
        - name: data
          emptyDir: {}
      restartPolicy: Never
```

## TODO
- [ ] Create a 'wait'-like command that watches a Job, streams its logs
and then downloads all the files from the Job.