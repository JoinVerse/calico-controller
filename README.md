# calico-controller: Controller for Calico Policies as a Kubernetes Custom Resource

This project enables you to manage Calico network policies directly as a Kubernetes resource. This way, you can manager network polcies the same way you manage the rest of your infrastructure resources.

### Example of a CalicoPolicy

The following CalicoPolicy enables incoming connections from pods with tag `'app': 'myservice'` in the namespace `myservice` to pods with tag `'app': 'postgres'` in the namespace `postgres`.

```
apiVersion: calico.verse.me/v1
kind: CalicoPolicy
metadata:
  name: myservice-to-postgres
spec:
  ingress:
  - action: allow
    source:
      selector: calico/k8s_ns == 'myservice' && app == 'myservice'
  order: 1000
  selector: calico/k8s_ns == 'postgres' && app == 'postgres'
```

### Integration with kubectl  

Using kubectl to apply the policy:
```
$ kubectl apply -f myservice-to-postgres.yaml
```

Use kubectl to get info about your policy:
```
$ kubectl describe cp myservice-to-postgres
Name:         myservice-to-postgres
Namespace:    
Labels:       <none>
Kind:         CalicoPolicy
Metadata:
  Cluster Name:                   
  Creation Timestamp:             2017-09-07T01:21:17Z
  Deletion Grace Period Seconds:  <nil>
  Deletion Timestamp:             <nil>
  Generation:                     0
  Resource Version:               13552495
  Self Link:                      /apis/calico.verse.me/v1/calicopolicies/myservice-to-postgres
  UID:                            d8c50fdd-936a-11e7-9a1b-42010a84000a
Spec:
  Ingress:
    Action:  allow
    Source:
      Selector:  calico/k8s_ns == 'myservice' && app == 'myservice'
  Order:         1000
  Selector:      calico/k8s_ns == 'postgres' && app == 'postgres'
Events:          <none>
```

  
