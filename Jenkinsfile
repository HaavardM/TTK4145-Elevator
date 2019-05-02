
def label = "worker-${UUID.randomUUID().toString()}"

podTemplate(label: label, containers: [
    containerTemplate(name: 'docker', image: 'docker', command: 'cat', ttyEnabled: true)
],
volumes: [
    hostPathVolume(mountPath: '/var/run/docker.sock', hostPath: '/var/run/docker.sock')
]) {

    node(label) {
        def repo = checkout scm
        def commit = repo.GIT_COMMIT
        def branch = repo.GIT_BRANCH
        def registry = "haavardm/ttk4145-elevator"
        def credentidal = "dockerhub"

        stage('Build image') {
            container('docker') {
                withCredentials([[$class: 'UsernamePasswordMultiBinding',
                credentialsId: 'dockerhub',
                usernameVariable: 'DOCKER_HUB_USER',
                passwordVariable: 'DOCKER_HUB_PASSWORD']]) {
                    sh """
                    docker login -u ${DOCKER_HUB_USER} -p ${DOCKER_HUB_PASSWORD}
                    docker build -t ${registry}:${commit} .
                    docker push ${registry}:${commit}
                    """
                }
            }
        }
    }
}
