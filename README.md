# Kubernetes AAD MSI

Authenticate to resources secured by Azure Active Directory (AAD) using Managed Service Identities (MSI) directly from Kubernetes.

## What problem does this solve?

Authentication is a difficult problem, and even in a cloud-first/cloud-native world it is still a tough problem to solve.

A feature in Azure that makes this a much easier problem to approach is Managed Service Identities. This allows Azure resources to automatically have an identity that can be used to authenticate against resources secured with Azure Active Directory (databases, storage, etc.).

Instead of passing around usernames and passwords or having to worry about baking in private keys to images, MSIs give us a very simple out-of-the-box experience that is secure and requires a lot less development effort.

Traditionally MSIs have been largely implemented directly from Virtual Machines (IaaS). In the Kubernetes world, we have an extra layer on top of VMs. But the usage of MSIs is still possible through the [aad-pod-identity](https://github.com/Azure/aad-pod-identity) project. For more information on exactly how it works under the covers, see the source repo for documentation.

## Example

In this repo I use the example of my application (living in a pod) that needs to access a resource in Azure. In my sample, I'm using an Azure SQL database.

## Steps

#### Create the AKS cluster

```
$ az group create -n resource_group -l eastus
$ az aks create -n k8scluster -g resource_group --node-count 1
$ az aks get-credentials -g resource_group -n k8scluster
```

#### Create and configure the Azure SQL server and database

```
$ az sql server create -g resource_group -n sql_server_name --admin-user admin_user --admin-password '<password>'
$ az sql db create -n testdb --server sql_server_name -g resource_group
```

Then you will need to set the Active Directory admin to be able to enable this feature for AAD auth against SQL.

You will also possibly need to configure your firewall on the SQL server to allow your client connections.

#### Create the aad-pod-identity resources

This is what does all of the handling for this in the Kubernetes cluster.

```
$ kubectl apply -f https://raw.githubusercontent.com/Azure/aad-pod-identity/master/deploy/infra/deployment-rbac.yaml
```

#### Create the managed identity that will be used for the pod(s)

```
$ az identity create -g $(az aks show -n k8scluster -g resource_group --query "nodeResourceGroup" -o tsv) -n k8scluster -o json
```

Save the output from this command, as we'll be needing the `clientId` and `id` data.


#### Create the AzureIdentity and AzureIdentityBinding resources

```
$ cat << EOF > /tmp/aadidentity.yaml
apiVersion: "aadpodidentity.k8s.io/v1"
kind: AzureIdentity
metadata:
  name: sqlaad1
spec:
  type: 0
  ResourceID: <id_from_identity>
  ClientID: <client_id_from_identity>
EOF

$ kubectl apply -f /tmp/aadidentity.yaml

$ cat << EOF > /tmp/aadidentitybinding.yaml
apiVersion: "aadpodidentity.k8s.io/v1"
kind: AzureIdentityBinding
metadata:
  name: sqlaadbinding1
spec:
  AzureIdentity: sqlaad1
  Selector: sqlaad
EOF

$ kubectl apply -f /tmp/aadidentitybinding.yaml
```

#### Create the SQL user

Now in the Azure SQL database, create the user to link it up with this Azure AD identity.

```sql
CREATE USER [k8scluster] FROM EXTERNAL PROVIDER;
EXEC sp_addrolemember 'db_owner', 'k8scluster';
```

*I added the user to `db_owner` for this demo, but for a more secure configuration you should give your users the least amount of privileges required.*

#### Create the SQL table and some test data

```sql
CREATE TABLE messagelist
(
    id INT IDENTITY(1, 1),
    message_text NVARCHAR(128) 
);

INSERT INTO messagelist
VALUES ('my message');

INSERT INTO messagelist
VALUES ('new message');
```

#### Building and deploying the application

The `build_and_deploy.sh` script automates this, but step-by-step we would now need to:

1. Build the application (`go build`)
1. Build the docker image (`docker build`)
1. Create the Kubernetes pod (`kubectl apply`)

#### Observations and explanations

You should see that the Kubernetes application living in the pod is able to successfully query the database using the Managed Service Identity.

```
$ kubectl logs aadtest1
```
