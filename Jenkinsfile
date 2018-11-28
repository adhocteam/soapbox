pipeline {
  agent {
    label 'general'
  }
  stages {
    stage ("Build containers") {
      steps {
        sh '''
          set -e
          docker build --target builder -t soapboxbuilder .
          docker build -f web/Dockerfile -t soapboxrails web
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
    stage ("Run rails tests") {
      steps {
        sh '''
          set -e
          docker run soapboxrails rspec
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
