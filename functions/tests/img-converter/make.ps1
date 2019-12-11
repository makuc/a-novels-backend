param(
    # Script to execute
    [Parameter()]
    [string]
    $exeFunc
)

# BEGIN Config

$functionName = "convert"

# END Config

function serve {
  ng serve
}
function clean {
    Remove-Item -LiteralPath "bin" -Force -Recurse -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath "z-full.jpg" -Force -Recurse -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath "z-thumb.jpg" -Force -Recurse -ErrorAction SilentlyContinue
    Write-Host "Cleaned!"
}
function build {
    Write-Host "Building the function"
    #$env:GOOS="linux";
    #$env:GOARCH="amd64";
    go build -o "bin/$functionName.exe"
}
function run {
    build
    Write-Host "Running..."
    &"./bin/$functionName.exe"
    Write-Host "Done!"
}
function tidy {
    go mod init
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
&$exeFunc
