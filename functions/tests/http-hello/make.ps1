param(
    # Script to execute
    [Parameter()]
    [string]
    $exeFunc
)

# BEGIN Config

$functionName = "http-hello"
$entryPoint = "Function"
$envVariables = "NAME=none"

# END Config

function serve {
  ng serve
}

function clean {
    Remove-Item -LiteralPath "bin" -Force -Recurse
}

function build {
    Write-Host "Building the function"
    go build
}

function tidy {
    go mod tidy
}

function upgrade {
    go get -u
}

function test {
    Write-Host "Executing the tests"
    go test .
}

function deploy {
    Write-Host "Deploying the function..."
	gcloud functions deploy `
        $functionName `
        --set-env-vars $envVariables `
        --trigger-http `
        --entry-point $entryPoint `
        --runtime=go111 `
        --memory=128MB
}


# RUNS the COMMAND
Clear-Host
$env:GONOPROXY="*github.com/makuc"
&$exeFunc
