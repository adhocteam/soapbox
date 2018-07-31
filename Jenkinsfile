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
          set -e
          unformatted=$(gofmt -l $(find . -type f -name '*.go' -not -path './vendor/*'))
          [[ -z "$unformatted" ]] && exit 0
          echo "$unformatted"
          exit 1
        '''
      }
    }
    stage ("Run go tests") {
      steps {
        sh '''
          set -e
          go test ./...
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
