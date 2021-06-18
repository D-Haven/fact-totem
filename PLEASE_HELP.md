# We need lots of help

In order to make this ready for prime time, there are lots of areas that need attention:

* Lots of testing 
* Enhanced application authorization
* Horizontal scaling support
* Containerization and Kubernetes support
* Branch and Merge functionality (may be way future), and may not be able to handle generally

### Support for K8s Service Accounts
Enable validating kubernetes service account JWTs.  Requires some additional tooling and research to get the list of
RSA public keys to check the service account signature.  This would allow easier deployments so that microservices
call with the service account token they are running under and the permissions file is much easier to manage.  Our
JWT library supports this of course, but its a question of working with the K8s infrastructure to get the keys to verify
service account tokens.

### Replication across multiple nodes
DGraph is built on top of Badger, and so are a number of other things.  If there is support at the underlying Badger
level for clustering and horizontal scaling, we need to add support for that to Fact Totem.