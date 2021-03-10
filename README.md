# AWS Elastic Container Service (ECS)

## Introduction

### What is Amazon Elastic Container Service?
> Amazon ECS helps you easy to run, stop and manage containers on a cluster. <br>
The containers are defined in a task definition that you use to run individual tasks or tasks within a service. <br>
> Moreover, you can run the tasks and services on a serverless infrastructure that is managed by AWS Fargate. <br>
> Alternatively, for more control over your infrastructure, you can run the tasks and services on a cluster of EC2 instances that you manage.

### What is Fargate?
> Fargate is one of ECS launch type. <br>
> It helps you use ECS to run containers **without managing server** or cluster of EC2 instances. <br>
> Each Fargate task has its own **isolation boundary** and does not share the underlying kernel, CPU resources, memory resources, or elastic network interface with another task.

<img src="./images/intro_ec2_fargate.png" width="80%" height="80%"> <br>
- In ECS launch type, you can choose **Fargate** or **EC2**.

### Main Components
#### ★ Cluster
> ECS Cluster is a **grouping of tasks or services**. <br>
> If you use EC2 launch type to run tasks or services, cluster is also a **grouping of container instances**.

#### ★ Task Definition
- Like a blueprint of your container.
- Can define **multiple containers** in a task definition.
- The following are parameters you can set in the task definition.
    - Launch type
    - CPU & memory
    - Docker image
    - Docker network mode *(Introduce below)*
    - Logging configuration
    - Whether task should continue to run when container finish or fail
    - The command the container should run when it is started
    - Data volumes should be userd in the containers
    - IAM role
- **Network mode**
    1. awsvpc
        > The task is allocated its own **Elastic Network Interface(ENI)** and a **private IP address**. This gives the task to have the same networking property as EC2 instances. <br>
        > ★ If the launch type is **Fargate**, the only Network mode is **awsvpc**.
    2. bridge
        > The task utilizes Docker's built-in virtual network which runs inside each EC2 instance hosting the task.
    3. host
        > The task bypasses Docker's built-in virtual network and maps container ports directly to the ENI of the EC2 instance hosting the task. As a result, you can't run multiple instantiations of the same task on a single EC2 instance when port mappings are used.
    4. none
        > The task has no external network connectivity.

#### ★ Services
> ECS Services helps you to **run and maintain a specified number of instances** of a task definition in the cluster. <br>
> If any tasks fail or stop, ECS Services scheduler launches another instance of your task definition to replace it in order to maintain the desired number of tasks in the service. <br>
- Often used with **Load Balance**. ECS Services maintain the number of tasks and Load balancer distributes traffic across the tasks that are associated with the ECS Services.
- The following are services you can set in the ECS Services.
    - Deployment types
    - Load Balancing
    - Auto Scaling
    - Service Discovery *(Introduce below)*
- **Service Discovery**
    > ★ This is key point can that container communecate with each other. <br>
    > Service discovery uses AWS Cloud Map API to manage HTTP and DNS namespaces for the ECS Services.
    - Consists of the following components:
        - namespace
            - domain name, such as `example.com`
        - service
            - subdomain name, exist in namespace.
        - ECS health checks
    - The Create Service workflow in the ECS console only supports Service Discovery register **private DNS namespaces**.

### What is Cloud Map?
> Cloud Map is a cloud resource discovery service, allows you to register any application resources with custom names, and no need to manage resources manually. <br>
> With Cloud Map, you can create and maintain a map of the backend services and resources that your applications depend on, such as databases, queues, microservices, and cloud resources. <br>
- Cloud Map is tightly integrated with ECS. When new container tasks spin up or down, they automatically register with Cloud Map. 
- When a Cloud Map private DNS namespace is created, Route 53 private hosted zone will be created automatically.

---
## Scenario
In this Lab, We will build a **microservice** that can count the number of times you visit the website. <br>
Using **Golang** to create the website, and expose **8000** port to allow traffic to enter from this port. <br>
And using **Redis** as the database, which port is **6379** by default. <br>
We need to deploy the Web and Redis in **different container** and allow them to **communicate with each other**. <br>
I don’t want Web and Redis to be accessed directly by internet, the traffic can only come in from **ALB**.

## Use Case
- Build a **3-tier web architecture**. Put Front-end, Back-end, Database in different containers.

## Overview
<img src="./images/microservice_architecture.png" width="80%" height="80%">

#### How will we achieve this?
1. Create a VPC environment.
2. Prepare the Dockerfile and build it into Image.
3. Create ECR repository.
4. Create ECS Cluster and Task Definition.
5. Create Application Load Balancer.
6. Create Cloud Map, that website can use the domain name to communicate with database.
7. Add the Task to ECS Service.

## Prerequisites
- Basic understanding of AWS **VPC**, **ALB**, and **Container** knowledge.
- Make sure the region is **US East (N. Virginia)**.
- Download the source file and project:
    - [vpc.yaml](./prerequisites/vpc.yaml)
    - [microservice_with_ECS](./materials)

## Step by Step
### Step 1 : Setup VPC network
- Go to [CloudFormation](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/)
- Use **CloudFormation** template ([vpc.yaml](./prerequisites/vpc.yaml)) to setup VPC network.
- We automatically generate the **Public Subnet x2**, **Private Subnet x2**, **Security Group x3**.

### Step 2 : Use Cloud9 to push image to ECR
#### Step 2.1 : Create two ECRs
- Go to [ECR Repositories](https://console.aws.amazon.com/ecr/repositories?region=us-east-1).
- In **Visibility settings**, select **Private**. Only attach IAM policy can access the repository.
- In **Repository name**, type the name you like, one for **Web**, the other for **Redis**.
<img src="./images/create_ECR_web.png" width="80%" height="80%">
<img src="./images/create_ECR_redis.png" width="80%" height="80%">

#### Step 2.2 : Create Cloud9 Environment
- Go to [AWS Cloud9](https://console.aws.amazon.com/cloud9/home/product), click create environment
- Type the name you like.
- Let all setting **default**. Then **Create environment**.
<img src="./images/cloud9.png" width="80%" height="80%">

#### Step 2.3 : Modify EC2 IAM role
- Go to [EC2 Instances console](https://console.aws.amazon.com/ec2/v2/home?region=us-east-1#Instances:).
- Find the instance we create through Cloud9, right-click and find **Security**, then click **Modify IAM role**.
- In **Modify IAM role**, select **Allow EC2 Access ECR**.
<img src="./images/modify_IAM_role.png" width="80%" height="80%">
<img src="./images/allow_ec2_access_ecr.png" width="80%" height="80%">

#### Step 2.4 : Build images
- Go back to **Cloud9** IDE and **upload local files**. ([microservice_with_ECS](./materials))
- There will have **four files** in the left panel.
    <img src="./images/cloud9_upload_file.png" width="80%" height="80%">
- First, in the command-line interface, **login** to AWS.
```
$(aws ecr get-login --no-include-email --region us-east-1)
```
- Then, **build Dockerfile** into a **Web** image, and push it into **Web repository**.
    - You need to find **Repository name** and **URI** in the **ECR**, which you just created.
```
docker build -t web -f Dockerfile_web .
docker tag web:latest {Your Repo URI}.dkr.ecr.us-east-1.amazonaws.com/{Your Repo Name}:latest
docker push {Your Repo URI}.dkr.ecr.us-east-1.amazonaws.com/{Your Repo Name}:latest
```
> ```docker build -t {image name} -f {Dockerfile path} .```<br>
> If no argument `-f`, docker will automatically find the file named ```Dockerfile``` in this path.
- Finally, **build Dockerfile** into **Redis** image, and push it into **Redis repository**.
```
docker build -t redis -f Dockerfile_redis .
docker tag redis:latest {Your Repo URI}.dkr.ecr.us-east-1.amazonaws.com/{Your Repo Name}:latest
docker push {Your Repo URI}.dkr.ecr.us-east-1.amazonaws.com/{Your Repo Name}:latest
```
> **build -> tag -> push**, is a process of pushing local image to remote repository.

#### Step 2.5 : Check ECR
- Click into repository, you will find there has a image you just pushed from Cloud9.
    <img src="./images/check_ECR.png" width="80%" height="80%">

### Step 3 : Create ECS Cluster
- Go to [ECS console -> Cluster](https://us-east-1.console.aws.amazon.com/ecs/home?region=us-east-1#/clusters).
- Choose **Fargate** as a launch type.
    <img src="./images/create_ECS_cluster_1.png" width="80%" height="80%">
- In **Networking**, **do not Create VPC**. We already setup vpc network from CloudFormation.
    <img src="./images/create_ECS_cluster_2.png" width="80%" height="80%">

### Step 4 : Create two Task Definitions, one for web, the other for redis
- In ECS console, click **Task Definitions** in the left penal.
- Choose **Fargate** as a launch type.
    <img src="./images/create_ECS_task_1.png" width="80%" height="80%">
- In **Task Definition Name**, type **web-task**.
- In **Task Role**, select **ecsTaskExecutionRole**.
    > If the launch type is **Fargate**, Network mode is fixed to **awsvpc**.
    <img src="./images/create_ECS_task_2.png" width="80%" height="80%">
- In **Task size**, select Task **memory** and **CPU**.
    <img src="./images/create_ECS_task_3.png" width="80%" height="80%">
- In **Container Definitions**, click **Add Container**.
    - Find **web Image URI** from ECR.
    - In **Port mappings**, type **8000** as expose port.
        > You can choose paste **repository-url**, or **{image name}:{tag}** as an image. <br>
        > If you choose **repository-url**, when you push the same image into ECR, task will always select the image which **tag is the latest**. <br>
    <img src="./images/create_ECS_task_web_container.png" width="80%" height="80%">
    
#### ★★★ Repeat the above steps to create redis Task Definition
- **★ Note ★**, In **Add Container**, 
    - Find **redis Image URI** from ECR.
    - In **Port mappings**, type **6379** as expose port.
        <img src="./images/create_ECS_task_redis_container.png" width="80%" height="80%">
> In Container Definitions, **Storage and Logging** has already set up **CloudWatch** by default. <br>
> If you run container has any problem or bug, can go to cloudwatch to check logs. <br>
> Also, you can setup the file you want to **mount** between local and container in this part.

### Step 5 : Create Service - redis
- Go back to **web-cluster** we created. In the **Services**, click **Create**.
    <img src="./images/create_service.png" width="80%" height="80%">
- In **Configure Service**
    - In **Launch type**, choose **Fargate**.
    - In **Task Definition,** select **redis-task**.
    - In **Service name**, type **redis**.
    - In **Number of tasks**, type **1**.
    <img src="./images/ECS_service_redis_1.png" width="80%" height="80%">
- In **Configure network**
    - In **Cluster VPC**, select **Web VPC** we create from CloudFormation.
    - In **Subnets**, we put container into **Private Subnet**.
    - In **Security group**, click **Edit**, and select **redis-sg**, which allows traffic enter from 6379 port.
    - In **Auto-assign public IP**, select **DISABLED**.
        <img src="./images/ECS_service_redis_2.png" width="80%" height="80%">
        <img src="./images/ECS_service_redis_3.png" width="80%" height="80%">
- We don't need any load balancer for redis service.
- In **Service discovery**
    - Create **Namespace**, type **internal-countweb.com**.
    - Type **redis** as service name.
        <img src="./images/service_discovery_1.png" width="80%" height="80%">
        <img src="./images/service_discovery_2.png" width="80%" height="80%">
- After create redis service which with Service Discovery, go to **Cloud Map**, you can find the **Domain name** you created.
- Click in, you can see there has one service **"redis"**.
    > **★ Note ★**  <br>
    > You need to use this Service name & Domain name to **connect** web and redis. <br>
    > You can find **"redis.internal-countweb.com"** is setting in the code of **main.go**, allowing web and redis to communicate. <br>
    > If you named the different name, please modify the variable **RedisEndpoint** which in **main.go**, and run the ECR process of **"build -> tag -> push"**. <br>
    > After pushing the new image, the tasks in the cluster will automatically restart by service scheduler.
    <img src="./images/cloudmap.png" width="80%" height="80%">

### Step 6 : Create Service - web
#### Step 6.1 : Create Application Load Balancer for web
> In this section, we let Load Balancer **listen port 80** and forward all traffic **from port 80 to port 8000** of the web. <br>
> So it is importent to set right **Listen (80)** and **Target Group (8000)**.

- Go to [EC2 -> Load Balancers](https://console.aws.amazon.com/ec2/v2/home?region=us-east-1#LoadBalancers:sort=loadBalancerName)
- In **load balancer type**, choose **HTTP/HTTPS**.
- In **Basic Configuration**
    - In **Name**, type **web-lb**.
    - In **Scheme**, choose **internet-facing**.
        > **Internet-facing routes** : requests from clients over the **Internet** to targets. <br>
        > **Internal routes** : requests from clients to targets using **private IP addresses**.
    - In **Listeners**, type **80** Port.
    - In **Availability Zones**, choose **Web VPC** and **Public Subnet**. So that you can use public DNS name to access website.
        <img src="./images/LB_1.png" width="80%" height="80%">
- In **Configure Security Groups**
    - Choose **http-sg**, which allows traffic enter from 80 & 443 port.
        <img src="./images/LB_2.png" width="80%" height="80%">
- In **Target Group**
    > When the traffic enter, Target Group control which resource should be redirected to.
    - In **Target Type**, select **IP**.
    - In **Port**, type **8000**, which is web's port.
- In **Health checks**, type **/home** as path. ALB will ping this URL to check website is healthy or not.
        <img src="./images/LB_3.png" width="80%" height="80%">

#### Step 6.2 : Go back to ECS to create service for web.
- In **Configure Service**
    - In **Launch type**, choose **Fargate**.
    - In **Task Definition,** select **web-task**.
    - In **Service name**, type **web**.
    - In **Number of tasks**, type **2**.
        <img src="./images/ECS_service_web_1.png" width="80%" height="80%">
- In **Configure network**
    - In **Cluster VPC**, select **Web VPC** we create from CloudFormation.
    - In **Subnets**, we put container into **Private Subnet**.
    - In **Security group**, click **Edit**, and select **golang-sg**, which allows traffic enter from 8000 port.
    - In **Auto-assign public IP**, select **DISABLED**.
        <img src="./images/ECS_service_web_2.png" width="80%" height="80%">
        <img src="./images/ECS_service_web_3.png" width="80%" height="80%">
- In **Container to load balance**
    - Click **Add to load balancer**.
    - In **Load Balancer name**, select **web-lb**.
    - In **Target group name**, choose **web-tg**, which created from ALB.
        <img src="./images/ECS_service_web_4.png" width="80%" height="80%">
        <img src="./images/ECS_service_web_5.png" width="80%" height="80%">

### Step 7 : Verify the Service is Success
- Wait for **Last status** of **tasks** on ECS cluster change to **"RUNNING"**.
- Go back to **EC2 -> Load Balancers**, select the load balancer you created.
- Copy **DNS name**, and paste to browser, you will see your IP and the number of visits to the website.
        <img src="./images/result.png" width="80%" height="80%">

## Conclusion
Congratulations!! <br>
You have learned how to use AWS ECS to create containers, and allow containers to communicate with each other.<br>
Next time you want to deploy a website, there is a lightweight way you can choose, run on the container instead of the VM.<br>

## Reference
[AWS ECS Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/Welcome.html) <br>
[AWS Cloud Map Developer Guide](https://docs.aws.amazon.com/cloud-map/latest/dg/what-is-cloud-map.html)