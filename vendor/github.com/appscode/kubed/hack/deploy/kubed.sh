#!/bin/bash
set -ou pipefail

echo "checking kubeconfig context"
kubectl config current-context || { echo "Set a context (kubectl use-context <context>) out of the following:"; echo; kubectl config get-contexts; exit 1; }
echo ""

# ref: https://stackoverflow.com/a/27776822/244009
case "$(uname -s)" in
    Darwin)
        curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.1.0/onessl-darwin-amd64
        chmod +x onessl
        export ONESSL=./onessl
        ;;

    Linux)
        curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.1.0/onessl-linux-amd64
        chmod +x onessl
        export ONESSL=./onessl
        ;;

    CYGWIN*|MINGW32*|MSYS*)
        curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.1.0/onessl-windows-amd64.exe
        chmod +x onessl.exe
        export ONESSL=./onessl.exe
        ;;
    *)
        echo 'other OS'
        ;;
esac

# http://redsymbol.net/articles/bash-exit-traps/
function cleanup {
    rm -rf $ONESSL ca.crt ca.key server.crt server.key
}
trap cleanup EXIT

# ref: https://stackoverflow.com/a/7069755/244009
# ref: https://jonalmeida.com/posts/2013/05/26/different-ways-to-implement-flags-in-bash/
# ref: http://tldp.org/LDP/abs/html/comparison-ops.html

export KUBED_NAMESPACE=kube-system
export KUBED_SERVICE_ACCOUNT=default
export KUBED_ENABLE_RBAC=false
export KUBED_RUN_ON_MASTER=0
export KUBED_DOCKER_REGISTRY=appscode
export KUBED_IMAGE_PULL_SECRET=
export KUBED_UNINSTALL=0

show_help() {
    echo "kubed.sh - install Kubernetes cluster daemon"
    echo " "
    echo "kubed.sh [options]"
    echo " "
    echo "options:"
    echo "-h, --help                         show brief help"
    echo "-n, --namespace=NAMESPACE          specify namespace (default: kube-system)"
    echo "    --rbac                         create RBAC roles and bindings"
    echo "    --docker-registry              docker registry used to pull kubed images (default: appscode)"
    echo "    --image-pull-secret            name of secret used to pull kubed operator images"
    echo "    --run-on-master                run kubed operator on master"
    echo "    --uninstall                    uninstall kubed"
}

while test $# -gt 0; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        -n)
            shift
            if test $# -gt 0; then
                export KUBED_NAMESPACE=$1
            else
                echo "no namespace specified"
                exit 1
            fi
            shift
            ;;
        --namespace*)
            export KUBED_NAMESPACE=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
        --docker-registry*)
            export KUBED_DOCKER_REGISTRY=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
        --image-pull-secret*)
            secret=`echo $1 | sed -e 's/^[^=]*=//g'`
            export KUBED_IMAGE_PULL_SECRET="name: '$secret'"
            shift
            ;;
        --rbac)
            export KUBED_SERVICE_ACCOUNT=kubed-operator
            export KUBED_ENABLE_RBAC=true
            shift
            ;;
        --run-on-master)
            export KUBED_RUN_ON_MASTER=1
            shift
            ;;
        --uninstall)
            export KUBED_UNINSTALL=1
            shift
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
done

if [ "$KUBED_UNINSTALL" -eq 1 ]; then
    kubectl delete deployment -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete service -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete secret -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete apiservice -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete validatingwebhookconfiguration -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete mutatingwebhookconfiguration -l app=kubed --namespace $KUBED_NAMESPACE
    # Delete RBAC objects, if --rbac flag was used.
    kubectl delete serviceaccount -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete clusterrolebindings -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete clusterrole -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete rolebindings -l app=kubed --namespace $KUBED_NAMESPACE
    kubectl delete role -l app=kubed --namespace $KUBED_NAMESPACE

    exit 0
fi

env | sort | grep KUBED*
echo ""

# create necessary TLS certificates:
# - a local CA key and cert
# - a webhook server key and cert signed by the local CA
$ONESSL create ca-cert
$ONESSL create server-cert server --domains=kubed-operator.$KUBED_NAMESPACE.svc
export SERVICE_SERVING_CERT_CA=$(cat ca.crt | $ONESSL base64)
export TLS_SERVING_CERT=$(cat server.crt | $ONESSL base64)
export TLS_SERVING_KEY=$(cat server.key | $ONESSL base64)
export KUBE_CA=$($ONESSL get kube-ca | $ONESSL base64)

kubectl get secret kubed-config -n kube-system > /dev/null 2>&1
if [ "$?" -eq 1 ]; then
    kubectl create secret generic kubed-config -n $KUBED_NAMESPACE \
        --from-literal=config.yaml=$(curl -fsSL https://raw.githubusercontent.com/appscode/kubed/0.6.0-rc.0/hack/deploy/config.yaml)
    kubectl label secret kubed-config app=kubed -n $KUBED_NAMESPACE
fi

curl -fsSL https://raw.githubusercontent.com/appscode/kubed/0.6.0-rc.0/hack/deploy/operator.yaml | $ONESSL envsubst | kubectl apply -f -

if [ "$KUBED_ENABLE_RBAC" = true ]; then
    kubectl create serviceaccount $KUBED_SERVICE_ACCOUNT --namespace $KUBED_NAMESPACE
    kubectl label serviceaccount $KUBED_SERVICE_ACCOUNT app=kubed --namespace $KUBED_NAMESPACE
    curl -fsSL https://raw.githubusercontent.com/appscode/kubed/0.6.0-rc.0/hack/deploy/rbac-list.yaml | $ONESSL envsubst | kubectl auth reconcile -f -
    curl -fsSL https://raw.githubusercontent.com/appscode/kubed/0.6.0-rc.0/hack/deploy/user-roles.yaml | $ONESSL envsubst | kubectl auth reconcile -f -
fi

if [ "$KUBED_RUN_ON_MASTER" -eq 1 ]; then
    kubectl patch deploy kubed-operator -n $KUBED_NAMESPACE \
      --patch="$(curl -fsSL https://raw.githubusercontent.com/appscode/kubed/0.6.0-rc.0/hack/deploy/run-on-master.yaml)"
fi
