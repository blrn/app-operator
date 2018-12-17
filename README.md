# app-operator
---
## Example
The following Application will run [Ghost](https://github.com/TryGhost/Ghost) and create a service and ingress.
```yaml
apiVersion: apps.blrn.io/v1alpha1
kind: Application
metadata:
  name: example-application
  namespace: app-operator
spec:
  # Add fields here
  image: ghost:2-alpine
  service:
    port: 80
    targetPort: 2368
  ingress:
    host: test-operator.blrn.io
    targetPort: 80
```

## TODO
- Persistent Volumes
- Envoy sidecar injection
- Don't require service to be specified if ingress is specified. 
- ???
