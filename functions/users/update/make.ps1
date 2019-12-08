param(
    # Script to execute
    [Parameter()]
    [string]
    $exeFunc
)

# BEGIN Config

$functionName = "user-update"
$entryPoint = "UserUpdate"
$projectId = "testing-192515"
$triggerEvent = "providers/cloud.firestore/eventTypes/document.update"
$triggerResource = "projects/$projectId/databases/(default)/documents/users/{uid}"
$envVariables = "worker_id=full-admin-rights,leaseSeconds=60"

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

function test {
    Write-Host "Executing the tests"
    go test .
}

function deploy {
    Write-Host "Deploying the function..."
	gcloud functions deploy `
        $functionName `
        --set-env-vars $envVariables `
        --trigger-event $triggerEvent `
        --trigger-resource $triggerResource `
        --entry-point $entryPoint `
        --runtime=go111 `
        --memory=128MB
}


# RUNS the COMMAND
Clear-Host
$env:GOPRIVATE="*github.com/makuc"
&$exeFunc
