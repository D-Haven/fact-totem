[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=D-Haven_fact-totem&metric=alert_status)](https://sonarcloud.io/dashboard?id=D-Haven_fact-totem)

# Fact Totem
The name "Fact Totem" is layered:

* Facts are a central concept in this tool
* Totems have symbolic meaning
* Factotums are employees that do all kinds of work

While Fact Totem isn't an employee, everything else applies here.

# Key Concepts
Fact Totem is an event store designed to allow authenticated access to your event storage.  The naming conventions we
use will hopefully minimize confusion when adopting event sourced architectures:

* **Aggregate:** The type of thing you are tracking history on.  The name comes from Domain Driven Design aggregate roots
* **Entity:** The specific instance of an aggregate to record or retrieve facts
* **Fact:** The record of something that occurred on that entity.  The name fact helps to differentiate events used for
  inter-service communication, and those intended to be the current state of the entity
  
## Authorization
Fact Totem is designed to work within a larger application, and it leverages JWT tokens to negotiate the user and look
up permissions.  There is no built in user management with the tool, it simply uses a `permissions.yaml` file to map
permissions to users.  This allows you to inject the permissions file for your system as a ConfigMap in k8s.

The format of the permissions file is simply a list mapping "subject", and the list of aggregates for each subject.

```yaml
- subject: "match:jwt:sub"
  read: ["*"]
  append: ["*"]
  scan: ["*"]
- subject: "*"
  read: ["*"]
  scan: ["*"]
```

There are three permissions:

|Permission|Description|
|----------|-----------|
| Read | The subject is allowed to read any entity for the named list of aggregates (or `*` for all aggregates) |
| Append | The subject is allowed to append new facts to any entity in the named list of aggregates (or `*` for all aggregates) |
| Scan | The subject is allowed to scan for the list of all entities for the named list of aggregates (or `*` for all aggreates) |


## Under the covers
Fact Totem is built on top of the venerable [BadgerDb](https://github.com/dgraph-io/badger).  What you get for free from
approach includes:

* Command line tool to back up and restore the database
* Command line tool to rotate the master key