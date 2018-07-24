pipeline {
  agent {
    dockerfile {
      additionalBuildArgs '--target builder'
    }
  }
  stages {
    stage ("Check formatting") {
      steps {
        sh '''
          unformatted=$(gofmt -l $(find . -type f -name '*.go' -not -path './vendor/*'))
          [[ -z "$unformatted" ]] && exit 0
          echo "$unformatted"
          exit 1
        '''
      }
    }
  }

  post {
    always {
      deleteDir()
      cleanWs()
    }
  }
}
