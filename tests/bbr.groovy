/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
    label "${params.buildNode}"
  }

  stages {
    stage('Download BBR') {
      steps {
        sh """
          curl -s https://api.github.com/repos/opencord/bbsim/releases/latest | grep "browser_download_url.*bbr-linux-amd" | cut -d : -f 2,3 | tr -d \\" | wget -qi -
          mv bbr-linux-amd64 bbr
          chmod a+x bbr
        """
      }
    }
    stage('Download BBSim') {
      steps {
        sh """
          docker pull voltha/bbsim:master
        """
      }
    }
    stage('64 ONUs (16 ONU x 4 PONs)') {
      steps {
        sh """
            docker rm -f bbsim || true
        """
        timeout(1) {
          sh """
            DOCKER_REPOSITORY=voltha/ DOCKER_TAG=master DOCKER_RUN_ARGS="-logLevel warn -enableperf -pon 4 -onu 16" make docker-run
            sleep 5
            ./bbr -pon 4 -onu 16 -logfile bbr_16_4.logs
            docker logs bbsim 2>&1 | tee bbsim_16_4.logs
            res=\$(cat bbr_16_4.logs | grep Duration | awk '{print \$5}' ); printf "Runtime\n\${res:9:10}" > bbr_16_4.txt
          """
        }
      }
    }
    stage('128 ONUs (32 ONU x 4 PONs)') {
      steps {
        sh """
            docker rm -f bbsim || true
        """
        timeout(1) {
          sh """
            DOCKER_REPOSITORY=voltha/ DOCKER_TAG=master DOCKER_RUN_ARGS="-logLevel warn -enableperf -pon 4 -onu 32" make docker-run
            sleep 5
            ./bbr -pon 4 -onu 32 -logfile bbr_32_4.logs
            docker logs bbsim 2>&1 | tee bbsim_32_4.logs
            res=\$(cat bbr_32_4.logs | grep Duration | awk '{print \$5}' ); printf "Runtime\n\${res:9:10}" > bbr_32_4.txt
          """
        }
      }
    }
    stage('256 ONUs (32 ONU x 8 PONs)') {
      steps {
        sh """
            docker rm -f bbsim || true
        """
        timeout(1) {
          sh """
            DOCKER_REPOSITORY=voltha/ DOCKER_TAG=master DOCKER_RUN_ARGS="-logLevel warn -enableperf -pon 8 -onu 32" make docker-run
            sleep 5
            ./bbr -pon 8 -onu 32 -logfile bbr_32_8.logs
            docker logs bbsim 2>&1 | tee bbsim_32_8.logs
            res=\$(cat bbr_32_8.logs | grep Duration | awk '{print \$5}' ); printf "Runtime\n\${res:9:10}" > bbr_32_8.txt
          """
        }
      }
    }
    stage('512 ONUs (64 ONU x 8 PONs)') {
      steps {
        sh """
            docker rm -f bbsim || true
        """
        timeout(1) {
          sh """
            DOCKER_REPOSITORY=voltha/ DOCKER_TAG=master DOCKER_RUN_ARGS="-logLevel warn -enableperf -pon 8 -onu 64" make docker-run
            sleep 5
            ./bbr -pon 8 -onu 64 -logfile bbr_64_8.logs
            docker logs bbsim 2>&1 | tee bbsim_64_8.logs
            res=\$(cat bbr_64_8.logs | grep Duration | awk '{print \$5}' ); printf "Runtime\n\${res:9:10}" > bbr_64_8.txt
          """
        }
      }
    }
    stage('960 ONUs (64 ONU x 15 PONs)') {
      steps {
        sh """
            docker rm -f bbsim || true
        """
        timeout(1) {
          sh """
            DOCKER_REPOSITORY=voltha/ DOCKER_TAG=master DOCKER_RUN_ARGS="-logLevel warn -enableperf -pon 15 -onu 64" make docker-run
            sleep 5
            ./bbr -pon 15 -onu 64 -logfile bbr_64_15.logs
            docker logs bbsim 2>&1 | tee bbsim_64_15.logs
            res=\$(cat bbr_64_15.logs | grep Duration | awk '{print \$5}' ); printf "Runtime\n\${res:9:10}" > bbr_64_15.txt
          """
        }
      }
    }
  }
  post {
    always {
      archiveArtifacts artifacts: '*.logs'
      step([$class: 'Mailer', notifyEveryUnstableBuild: true, recipients: "joey@opennetworking.org", sendToIndividuals: false])
    }
    success {
      plot([
        csvFileName: 'plot-bbr_16_4.csv',
        csvSeries: [[file: 'bbr_16_4.txt']],
        group: 'BBSim',
        title: '64 ONUs (16 ONUs x 4 PONs)',
        style: 'line'
      ])
      plot([
        csvFileName: 'plot-bbr_32_4.csv',
        csvSeries: [[file: 'bbr_32_4.txt']],
        group: 'BBSim',
        title: '128 ONUs (32 ONUs x 4 PONs)',
        style: 'line'
      ])
      plot([
        csvFileName: 'plot-bbr_32_8.csv',
        csvSeries: [[file: 'bbr_32_8.txt']],
        group: 'BBSim',
        title: '256 ONUs (32 ONUs x 8 PONs)',
        style: 'line'
      ])
      plot([
        csvFileName: 'plot-bbr_64_8.csv',
        csvSeries: [[file: 'bbr_64_8.txt']],
        group: 'BBSim',
        title: '512 ONUs (64 ONUs x 8 PONs)',
        style: 'line'
      ])
      plot([
        csvFileName: 'plot-bbr_64_15.csv',
        csvSeries: [[file: 'bbr_64_15.txt']],
        group: 'BBSim',
        title: '960 ONUs (64 ONUs x 15 PONs)',
        style: 'line'
      ])
    }
    failure {
      sh '''
      docker logs bbsim 2>&1 | tee bbsim_failed.logs
      docker cp bbsim:/var/lib/dhcp/dhcpd.leases dhcpd_leases_failed.logs || true
      docker exec bbsim bbsimctl onu list > onu_list.logs || true
      '''
    }
  }
}
