** **

# Project Description 

## 1. Vision and Goals Of The Project:

The two related prior projects we focus here are TiDB and Kubernetes. TiDB is a HTAP relational database that is readily available and can be easily scaled out. Kerbernetes is the default cloud OS.

The goal for this project is to create an operator that manages the lifecycle of TiDB clusters on Kubernetes. This operator aims to provide TiDB cluster management on Kubernetes including functions like starting, pausing and scaling out/in clusters in a declarative way. We also plan to build an observability ecosystem for the operator, including metrics, logging, tracing, and alerting.

## 2. Users/Personas Of The Project:

This section describes the principal user roles of the project together with the key characteristics of these roles. This information will inform the design and the user scenarios. A complete set of roles helps in ensuring that high-level requirements can be identified in the product backlog.

Again, the description should be specific enough that you can determine whether user A, performing action B, is a member of the set of users the project is designed for.

** **

## 3.   Scope and Features Of The Project:

The features of this project will include:

    Starting and pausing cluster - in scope
    Scaling in or out of cluster - in scope
    Monitoring and logging cluster - in scope
    Backup and restoring cluster - not in scope

In this project, a separate controller will be created for each feature. Each controller will correspond to a specific CRD.


## 4. Solution Concept
<!--
This section provides a high-level outline of the solution.

Global Architectural Structure Of the Project:

This section provides a high-level architecture or a conceptual diagram showing the scope of the solution. If wireframes or visuals have already been done, this section could also be used to show how the intended solution will look. This section also provides a walkthrough explanation of the architectural structure.

 

Design Implications and Discussion:

This section discusses the implications and reasons of the design decisions made during the global architecture design.
-->
TiDB will be deployed on Kubernetes. To provide automatic scaling, upgrading, monitoring and self-healing, an operator will be created and integrated with the cluster to operate on TiDB CRD. The operator will provide full life cycle management, to provide scalability and high availability.

## 5. Acceptance criteria

This section discusses the minimum acceptance criteria at the end of the project and stretch goals.

## 6.  Release Planning:

We will be using SCRUM for developing this project. Every group member will first get familiar with the project and then we will decide the detail of releasing plan next week.

<!-- Release planning section describes how the project will deliver incremental sets of features and functions in a series of releases to completion. Identification of user stories associated with iterations that will ease/guide sprint planning sessions is encouraged. Higher level details for the first iteration is expected. -->


