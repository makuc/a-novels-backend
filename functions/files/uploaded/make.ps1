param(
    # Script to execute
    [Parameter()]
    [string]
    $exeFunc
)

# BEGIN Config

$functionName = "on-file-uploaded"
$entryPoint = "OnFileUploaded"
$projectId = "testing-192515"
$triggerEvent = "google.storage.object.finalize"
$triggerResource = "$projectId.appspot.com"
$envVars = "WorkerID=full-admin-rights,leaseSeconds=60"

# END Config

function serve {
  ng serve
}
function clean {
    Write-Host "Clean up..."
    Remove-Item -LiteralPath "bin" -Force -Recurse -erroraction 'silentlycontinue'
    Write-Host "Done!"
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
        --trigger-event $triggerEvent `
        --trigger-resource $triggerResource `
        --entry-point $entryPoint `
        --set-env-vars $envVars `
        --runtime=go111 `
        --memory=128MB
}


# RUNS the COMMAND
Clear-Host
$env:GONOPROXY="*github.com/makuc"
&$exeFunc
