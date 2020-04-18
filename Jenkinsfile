pipeline {
    agent { label 'master' }
    options {
        ansiColor('xterm')
    }

    stages {
        stage('Checkout') {
            steps {
                script {
                GIT_REVISION = sh(returnStdout: true, script: """
                    git rev-parse --short HEAD"""
                ).trim()
                def PWD = pwd()
                echo "Check out REVISION: $GIT_REVISION on $PWD"
                PUSH_DOCKER_IMAGE_AND_DEPLOY_TO_INT = (['master', 'jenkins'].contains(GIT_BRANCH) ||
                    GIT_BRANCH ==~ /release\-[\d\-\.]+/ ||
                    GIT_BRANCH ==~ /[^\s]+enable_docker_image_push$/)
                checkout changelog: false, poll: false, scm: [$class: 'GitSCM', branches: [[name: '*/master']], doGenerateSubmoduleConfigurations: false, extensions: [[$class: 'RelativeTargetDirectory', relativeTargetDir: 'jenkins-helper']], submoduleCfg: [], userRemoteConfigs: [[credentialsId: 'github-personal-jenkins', url: 'https://github.com/sunshine69/jenkins-helper.git']]]
                utils = load("${WORKSPACE}/jenkins-helper/deployment.groovy")
                }//script
            }//steps
        }//stage

        stage('Build') {
            steps {
                script {
                    echo "Start build"
                    VERSION_PREFIX = "${GIT_BRANCH}-${GIT_REVISION}-".replace('/', '-')
                    BUILD_VERSION = VersionNumber projectStartDate: '2018-11-07', versionNumberString: "${BUILD_NUMBER}", versionPrefix: "${VERSION_PREFIX}", worstResultForIncrement: 'SUCCESS'
                    echo "Version:  $BUILD_VERSION"
                    echo "Revision: $GIT_REVISION"
                    sh './build-static.sh'
                }
            }
        }
        stage('Gather artifacts') {
            steps {
                script {
                    utils.save_build_data(['artifact_class': 'nsre'])

                    if (PUSH_DOCKER_IMAGE_AND_DEPLOY_TO_INT) {
                      archiveArtifacts allowEmptyArchive: true, artifacts: 'nsre-linux-amd64-static', fingerprint: true, onlyIfSuccessful: true
                    }
                    else {
                      echo "Not collecting artifacts as branch does not start with 'release' or branch is not 'develop', 'jenkins'"
                    }// If GIT_BRANCH
                } //script
            }
        }// Gather artifacts

    }
    post {
        always {
            script {
                utils.apply_maintenance_policy_per_branch()
                currentBuild.description = """Artifact version: ${BUILD_VERSION}<br/>
Artifact revision: ${GIT_REVISION}"""
            }
        }
        success {
            script {
              cleanWs cleanWhenFailure: false, cleanWhenNotBuilt: false, cleanWhenUnstable: false, deleteDirs: true, disableDeferredWipeout: true, notFailBuild: true, patterns: [[pattern: 'PerformanceTesting', type: 'INCLUDE'], [pattern: '*', type: 'INCLUDE']]
            }//script
        }
        failure {
            script {
                //slackSend baseUrl: 'https://xvt.slack.com/services/hooks/jenkins-ci/', botUser: true, channel: '#errcd-activity', message: "@here CRITICAL - ${JOB_NAME} (${BUILD_URL}) branch (${BRANCH_NAME}) revision (${GIT_REVISION}) on (${NODE_NAME})", teamDomain: 'xvt', tokenCredentialId: 'jenkins-ci-integration-token', color: "danger"
                echo 'Build failed'
            }//script
        }
    }
}
