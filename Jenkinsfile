pipeline {
  agent any
  stages {
    stage ("Build container") {
      steps {
        sh '''
          set -e
          docker build --target builder -t soapboxbuilder .
        '''
      }
    }
    stage ("Check formatting") {
      steps {
        sh '''
          set -e
          unformatted=$(docker run soapboxbuilder gofmt -l $(find . -type f -name '*.go' -not -path './vendor/*'))
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
          docker run soapboxbuilder go test ./...
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
