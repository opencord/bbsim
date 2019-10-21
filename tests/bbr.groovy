/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

 pipeline {

  /* no label, executor is determined by JJB */
  agent {
    label "${params.executorNode}"
  }

  stages {
    stage('Build BBSim') {
      steps {
        sh """
          make docker-build
        """
      }
    }
    stage('Build BBR') {
      steps {
        sh """
          make build-bbr
        """
      }
    }
    stage('64 ONUs (16 ONU x 4 PONs)') {
      steps {
        sh """
          DOCKER_RUN_ARGS="-pon 4 -onu 16" make docker-run
          ./bbr -pon 4 -onu 16 2>&1 | tee bbr_16_4.logs
          docker logs bbsim 2>&1 | tee bbsim_16_4.logs
        """
      }
    }
    post {
      always {
         archiveArtifacts artifacts: '*.logs'
         step([$class: 'Mailer', notifyEveryUnstableBuild: true, recipients: "teo@opennetworking.org", sendToIndividuals: false])
      }
      failure {
        sh '''
        docker logs bbsim 2>&1 | tee bbsim_failed.logs
        '''
      }
    }
  }
}
