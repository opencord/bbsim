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
    stage('Build BBR') {
      steps {
        sh """
          make build
        """
      }
    }
    stage('Build BBSim') {
      steps {
        sh """
          docker pull voltha/bbsim:master
          DOCKER_REPOSITORY=voltha/ DOCKER_TAG=candidate make docker-build
        """
      }
    }
    stage('64 ONUs (16 ONU x 4 PONs)') {
      steps {
        timeout(1) {
          sh """
            docker rm -f bbsim || true
            DOCKER_REPOSITORY=voltha/ DOCKER_TAG=candidate DOCKER_RUN_ARGS="-pon 4 -onu 16" make docker-run
            sleep 5
            ./bbr -pon 4 -onu 16
            docker logs bbsim 2>&1 | tee bbsim_16_4.logs
          """
        }
      }
    }
    stage('128 ONUs (32 ONU x 4 PONs)') {
      steps {
        timeout(1) {
          sh """
            docker rm -f bbsim || true
            DOCKER_REPOSITORY=voltha/ DOCKER_TAG=candidate DOCKER_RUN_ARGS="-pon 4 -onu 32" make docker-run
            sleep 5
            ./bbr -pon 4 -onu 32
            docker logs bbsim 2>&1 | tee bbsim_32_4.logs
          """
        }
      }
    }
    stage('256 ONUs (32 ONU x 8 PONs)') {
      steps {
        timeout(1) {
          sh """
            docker rm -f bbsim || true
            DOCKER_REPOSITORY=voltha/ DOCKER_TAG=candidate DOCKER_RUN_ARGS="-pon 8 -onu 32" make docker-run
            sleep 5
            ./bbr -pon 8 -onu 32
            docker logs bbsim 2>&1 | tee bbsim_32_8.logs
          """
        }
      }
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
      docker cp bbsim:/app/dhcp.logs dhcp_failed.logs
      docker cp bbsim:/var/lib/dhcp/dhcpd.leases dhcpd_leases_failed.logs
      docker cp bbsim:/app/tcpdump.logs tcpdump_failed.logs
      docker exec bbsim bbsimctl onu list > onu_list.logs
      '''
    }
  }
}
