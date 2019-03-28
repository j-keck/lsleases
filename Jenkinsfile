#!/usr/bin/env groovy

node("nix") {
    checkout scm
    stage("prepare") {

        stage("collect props") {
            env.LSLEASES_VERSION =  sh(script: "git describe --always --tags", returnStdout: true).trim()
        }

        stage("create manpage") {
            sh "nix-build -A manpage"
            dir("result") {
                stash name: "manpage", includes: "lsleases.1"
            }
        }
    }

}


parallel(
    "freebsd": {
        node("freebsd") {
            dir("build/freebsd") {
                env.LSLEASES_PACKAGE = "lsleases-${LSLEASES_VERSION}.txz"

                checkout scm
                build("freebsd", "amd64")

                unstash name: "manpage"
                sh "build-scripts/freebsd.sh"

                echo "sign package"
                sign(LSLEASES_PACKAGE, GPG_PASSPHRASE)

                echo "archive artifacts"
                archiveArtifacts artifacts: "${LSLEASES_PACKAGE}, ${LSLEASES_PACKAGE}.sig"
            }
        }
    },
    "dpkg": {
        node("nix") {
            dir("build/dpkg") {
                env.LSLEASES_PACKAGE = "lsleases-${LSLEASES_VERSION}.dpkg"

                checkout scm
                build("linux", "amd64")

                unstash name: "manpage"
                sh "nix-build -A package-dpkg"

                echo "sign package"
                sign(LSLEASES_PACKAGE, GPG_PASSPHRASE)

                echo "archive artifacts"
                archiveArtifacts artifacts: "${LSLEASES_PACKAGE}, ${LSLEASES_PACKAGE}.sig"
            }
        }
    }


        // },
        // "build@linux": {
        //     node() {
        //         prepare(REPO_URL)
        //         build("linux", "amd64")
        //         sh "file src/lsleases/lsleases"
        //     }
        // }
        // ,
        // "build@nix": {
        //     // for nixos, only verify it builds from 'default.nix'
        //     node("nix") {
        //         git REPO_URL
        //         sh "nix-build default.nix"
        //     }
        // },
)



def build(goos, goarch) {
    String goVersion = sh(script: "go version", returnStdout: true).trim()

    withEnv(["GO_EXTLINK_ENABLED=0", "CGO_ENABLED=0", "GOOS=${goos}", "GOARCH=${goarch}"]) {

        def goEnv = sh(script: "go env", returnStdout: true).trim().replaceAll("\n", ", ")

        echo "- build (${goVersion}, env: ${goEnv})"
        sh "go build -ldflags '-X main.VERSION=${LSLEASES_VERSION}' -v"
    }
}



def sign(String fileName, String passphrase) {
    String cmd = "gpg --pinentry-mode loopback --passphrase ${passphrase} " +
                 " --detach-sign --output ${fileName}.sig ${fileName}"

    echo "write signature in ${fileName}.sig"
    // execute without logging
    sh "#!/bin/sh -e\n${cmd}"
}
