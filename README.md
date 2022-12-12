** **

# Project Description 

## 1. Vision and Goals Of The Project:

The two related prior projects we focus here are TiDB and Kubernetes. TiDB is a HTAP relational database that is readily available and can be easily scaled out. Kerbernetes is the default cloud OS.

The goal for this project is to create an operator that manages the lifecycle of TiDB clusters on Kubernetes. This operator aims to provide TiDB cluster management on Kubernetes including functions like starting, pausing and scaling out/in clusters in a declarative way. We also plan to build an observability ecosystem for the operator, including metrics, logging, tracing, and alerting.

## 2. Users/Personas Of The Project:

The principle users of the project are people who manage the workflow and operate with TIDB clusters, where TiDB is a HTAP relational database easily scaled out as other NoSQL databases as well as provide support for data consistency and distributed transaction support. 

In this project, we will build an operator that manages the lifecycle of TiDB clusters on Kubernetes, which will allow users to manage their TiDB clusters, like starting the cluster, pausing the cluster, and scaling out/in the cluster in a declarative way. As well as build an observability ecosystem for the operator, including metrics, logging, tracing, and alerting. Using modern cloud-native o11y systems, like Prometheus, Grafana, Jaeger, and Loki.

** **

## 3.   Scope and Features Of The Project:

The features of this project will include:

    Starting and pausing cluster - in scope
    Scaling in or out of cluster - in scope
    Monitoring and logging cluster - in scope
    Backup and restoring cluster - in scope

In this project, a separate controller will be created for each of the three components. Each controller will correspond to a specific CRD.


## 4. Solution Concept
<!--
This section provides a high-level outline of the solution.

Global Architectural Structure Of the Project:

This section provides a high-level architecture or a conceptual diagram showing the scope of the solution. If wireframes or visuals have already been done, this section could also be used to show how the intended solution will look. This section also provides a walkthrough explanation of the architectural structure.

 

Design Implications and Discussion:

This section discusses the implications and reasons of the design decisions made during the global architecture design.
-->
TiDB will be deployed on Kubernetes. To provide automatic scaling, upgrading, monitoring and self-healing, an operator will be created and integrated with the cluster to operate on TiDB CRD. The operator will provide full life cycle management, to provide scalability and high availability. The operator will provide full life cycle management, to guarantee scalability and high availability. The architecture of our final deliverable is as follows:

![avatar](/pics/deliverable.png)

The deliverable contains three separate controllers in the operator, each responisble of controlling one of the three components. In the TiDB cluster managed by our operator, each component has a corresponding custom resource (CR). The three components communicate over a service, in order to enable fail-over. PD and TiKV are stateful components, therefore they are connected to persistent volumes (PV), in order to preserve data upon failure. External clients will query the cluster via a service that exposes the TiDB cluster. The Kubernetes cluster is also equipped with a monitoring stack composed of Prometheus and Grafana. This combination gives us a better insight of the status of the cluster, and makes it easier to trouble shoot.

Prerequisites to run the operator on a local Kubernetes cluster using Kind: [Kind](https://kind.sigs.k8s.io/), [Docker](https://www.docker.com/), [Go(1.17+)](https://go.dev/), [Kubectl](https://kubernetes.io/docs/tasks/tools/) and [MySQL](https://www.mysql.com/) client.
To start the cluster and the operator run:
```bash
cd multi-controller
./buildAndRun.sh
```

## 5. Acceptance criteria

This section discusses the acceptance criteria at the end of the project.
The files we submit will contain the below information.

```
 all the codes of the project s
 a document to describe the proposal
 a shell script to run the operator
```

The acceptance criteria is to finish below functions of operator 
```
starting / pausing / scaling out / in a cluster of TIiDB on kubernetes
simulating failure in all or part of your cluster to test its resilience
monitoring the pod running status
```
And we have implemented them all.

## 6.  Release Planning:

We will be using SCRUM for developing this project. At present, all the functions in the scope are implemented.

<!-- Release planning section describes how the project will deliver incremental sets of features and functions in a series of releases to completion. Identification of user stories associated with iterations that will ease/guide sprint planning sessions is encouraged. Higher level details for the first iteration is expected. -->


