docker build -t emeris/api-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .

kind load docker-image emeris/api-server --name emeris

 helm upgrade api-server \
        --install \
        --kube-context kind-emeris \
        --namespace emeris \
        --set imagePullPolicy=Always \
        ./helm