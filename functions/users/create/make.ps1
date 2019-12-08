param(
    # Script to execute
    [Parameter()]
    [string]
    $exeFunc
)

# BEGIN Config

$functionName = "on-user-create"
$entryPoint = "OnUserCreate"
$projectId = "testing-192515"
$triggerEvent = "providers/firebase.auth/eventTypes/user.create"

# END Config

function serve {
  ng serve
}
function clean {
    Remove-Item -LiteralPath "bin" -Force -Recurse
}
function build {
    Write-Host "Building the function"
    $env:GOOS="linux";
    $env:GOARCH="amd64";
    go build -o bin/$functionName
}
function test {
    Write-Host "Executing the tests"
    go test .
}
function deploy {
    Write-Host "Deploying the function..."
	gcloud functions deploy `
           $functionName `
           --trigger-event $triggerEvent `
           --trigger-resource $projectId `
           --entry-point $entryPoint `
           --runtime=go111 `
           --memory=128MB
}


# RUNS the COMMAND
Clear-Host
&$exeFunc
