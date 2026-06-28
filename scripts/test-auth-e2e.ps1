# AhorraApp — Pruebas E2E del backend auth
# Ejecutar: ./scripts/test-auth-e2e.ps1
# Requisitos: docker compose up -d (backend corriendo en :8080)

$BASE = "http://localhost:8080"
$PASS = 0
$FAIL = 0
$TOKEN = ""

function Step($name, $script) {
  Write-Host "`n=== $name ===" -ForegroundColor Cyan
  & $script
}

function Expect($cond, $msg) {
  if ($cond) { Write-Host "PASS: $msg" -ForegroundColor Green; $script:PASS++ }
  else { Write-Host "FAIL: $msg" -ForegroundColor Red; $script:FAIL++ }
}

Step "1. Health check" {
  try {
    $r = Invoke-WebRequest "$BASE/api/v1/health" -UseBasicParsing
    Expect ($r.StatusCode -eq 200) "GET /health returns 200"
  } catch { Expect $false "Health: $_" }
}

Step "2. Register nuevo usuario" {
  $body = @{ email="mariatest@ahorrapp.com"; password="Pass1234"; display_name="Maria Test" } | ConvertTo-Json
  try {
    $r = Invoke-WebRequest "$BASE/api/v1/auth/register" -Method Post -Body $body -ContentType "application/json" -UseBasicParsing
    $j = $r.Content | ConvertFrom-Json
    $script:TOKEN = $j.token
    Expect ($r.StatusCode -eq 201) "Register returns 201"
    Expect (-not [string]::IsNullOrEmpty($j.token)) "Token presente"
    Expect ($j.user.email -eq "mariatest@ahorrapp.com") "Email devuelto correcto"
  } catch { Expect $false "Register: $_" }
}

Step "3. Register duplicado falla con 409" {
  $body = @{ email="mariatest@ahorrapp.com"; password="Pass1234"; display_name="dup" } | ConvertTo-Json
  try {
    Invoke-WebRequest "$BASE/api/v1/auth/register" -Method Post -Body $body -ContentType "application/json" -UseBasicParsing | Out-Null
    Expect $false "Se esperaba 409, pero registro exitoso"
  } catch {
    Expect ($_.Exception.Response.StatusCode.value__ -eq 409) "Duplicado devuelve 409"
  }
}

Step "4. Login con credenciales correctas" {
  $body = @{ email="mariatest@ahorrapp.com"; password="Pass1234" } | ConvertTo-Json
  try {
    $r = Invoke-WebRequest "$BASE/api/v1/auth/login" -Method Post -Body $body -ContentType "application/json" -UseBasicParsing
    $j = $r.Content | ConvertFrom-Json
    $script:TOKEN = $j.token
    Expect ($r.StatusCode -eq 200) "Login returns 200"
    Expect (-not [string]::IsNullOrEmpty($j.token)) "Token presente"
  } catch { Expect $false "Login: $_" }
}

Step "5. Login con contrasena incorrecta falla 401" {
  $body = @{ email="mariatest@ahorrapp.com"; password="wrongpass" } | ConvertTo-Json
  try {
    Invoke-WebRequest "$BASE/api/v1/auth/login" -Method Post -Body $body -ContentType "application/json" -UseBasicParsing | Out-Null
    Expect $false "Se esperaba 401, login exitoso"
  } catch {
    Expect ($_.Exception.Response.StatusCode.value__ -eq 401) "Bad password returns 401"
  }
}

Step "6. GET /auth/me con token valido" {
  $h = @{ Authorization = "Bearer $script:TOKEN" }
  try {
    $r = Invoke-WebRequest "$BASE/api/v1/auth/me" -Headers $h -UseBasicParsing
    $j = $r.Content | ConvertFrom-Json
    Expect ($r.StatusCode -eq 200) "/auth/me returns 200"
    Expect ($j.email -eq "mariatest@ahorrapp.com") "Email coincide"
    Expect ($j.display_name -eq "Maria Test") "Display name coincide"
  } catch { Expect $false "/auth/me: $_" }
}

Step "7. GET /auth/me sin token falla 401" {
  try {
    Invoke-WebRequest "$BASE/api/v1/auth/me" -UseBasicParsing | Out-Null
    Expect $false "Se esperaba 401"
  } catch {
    Expect ($_.Exception.Response.StatusCode.value__ -eq 401) "Sin token retorna 401"
  }
}

Step "8. GET /users/me/points con token" {
  $h = @{ Authorization = "Bearer $script:TOKEN" }
  try {
    $r = Invoke-WebRequest "$BASE/api/v1/users/me/points" -Headers $h -UseBasicParsing
    $j = $r.Content | ConvertFrom-Json
    Expect ($r.StatusCode -eq 200) "/users/me/points returns 200"
    Expect ($j.total_points -ge 0) "total_points es numero >= 0"
    Expect ($j.recent_transactions -is [array]) "recent_transactions es array"
  } catch { Expect $false "/users/me/points: $_" }
}

Step "9. GET /ranking/products/search con query invalida 400" {
  $h = @{ Authorization = "Bearer $script:TOKEN" }
  try {
    Invoke-WebRequest "$BASE/api/v1/ranking/products/search" -Headers $h -UseBasicParsing | Out-Null
    Expect $false "Se esperaba 400"
  } catch {
    Expect ($_.Exception.Response.StatusCode.value__ -eq 400) "Sin query devuelve 400"
  }
}

Step "10. GET /ranking/products/search?q=arroz valido 200" {
  $h = @{ Authorization = "Bearer $script:TOKEN" }
  try {
    $r = Invoke-WebRequest "$BASE/api/v1/ranking/products/search?q=arroz" -Headers $h -UseBasicParsing
    $j = $r.Content | ConvertFrom-Json
    Expect ($r.StatusCode -eq 200) "Search returns 200"
    Expect ($j.results -is [array]) "results es array"
  } catch { Expect $false "Search: $_" }
}

Write-Host "`n=== Resumen ===" -ForegroundColor Yellow
Write-Host "PASS: $script:PASS" -ForegroundColor Green
Write-Host "FAIL: $script:FAIL" -ForegroundColor Red
if ($script:FAIL -gt 0) { exit 1 }