# Script de prueba rápida de la API
# Ejecutar después de: docker-compose up

$baseUrl = "http://localhost:8080"
$green = "`e[32m"
$red = "`e[31m"
$yellow = "`e[33m"
$reset = "`e[0m"

function Test-Endpoint {
    param($Method, $Path, $Body = $null, $Token = $null)
    
    $url = "$baseUrl$Path"
    $headers = @{}
    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }
    if ($Body) {
        $headers["Content-Type"] = "application/json"
    }

    try {
        if ($Method -eq "GET") {
            $response = Invoke-RestMethod -Uri $url -Method $Method -Headers $headers -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $url -Method $Method -Headers $headers -Body ($Body | ConvertTo-Json) -ErrorAction Stop
        }
        Write-Host "${green}✅ $Method $Path - OK${reset}"
        return $response
    } catch {
        Write-Host "${red}❌ $Method $Path - FAILED${reset}"
        Write-Host "   Error: $($_.Exception.Message)"
        return $null
    }
}

Write-Host "`n🧪 Testing Ergonomia API...`n" -ForegroundColor Cyan
Write-Host "Base URL: $baseUrl`n"

# 1. Health check
Write-Host "${yellow}--- Health Check ---${reset}"
Test-Endpoint -Method "GET" -Path "/health"

# 2. Listar productos (público)
Write-Host "`n${yellow}--- Public Endpoints ---${reset}"
$products = Test-Endpoint -Method "GET" -Path "/api/v1/products?limit=5"

# 3. Obtener producto específico
if ($products -and $products.products.Count -gt 0) {
    $slug = $products.products[0].slug
    Test-Endpoint -Method "GET" -Path "/api/v1/products/$slug"
}

# 4. Registro
Write-Host "`n${yellow}--- Auth Endpoints ---${reset}"
$registerBody = @{
    email = "apitest@example.com"
    password = "TestPassword123!"
    first_name = "API"
    last_name = "Test"
}
$register = Test-Endpoint -Method "POST" -Path "/api/v1/auth/register" -Body $registerBody

# Guardar token si el registro fue exitoso
$token = $null
if ($register -and $register.access_token) {
    $token = $register.access_token
    Write-Host "   Access token received: $($token.Substring(0,20))..."
} else {
    # Intentar login si el registro falló (usuario ya existe)
    Write-Host "   Trying login instead..."
    $loginBody = @{
        email = "apitest@example.com"
        password = "TestPassword123!"
    }
    $login = Test-Endpoint -Method "POST" -Path "/api/v1/auth/login" -Body $loginBody
    if ($login -and $login.access_token) {
        $token = $login.access_token
    }
}

# 5. Endpoints protegidos (si tenemos token)
if ($token) {
    Write-Host "`n${yellow}--- Protected Endpoints ---${reset}"
    
    # Obtener usuario actual
    $me = Test-Endpoint -Method "GET" -Path "/api/v1/me" -Token $token
    
    # Listar favoritos (vacío probablemente)
    Test-Endpoint -Method "GET" -Path "/api/v1/favorites" -Token $token
    
    # Crear comparación (si hay productos)
    if ($products -and $products.products.Count -ge 2) {
        $comparisonBody = @{
            title = "Test Comparison"
            product_ids = @($products.products[0].id, $products.products[1].id)
        }
        $comparison = Test-Endpoint -Method "POST" -Path "/api/v1/comparisons" -Body $comparisonBody -Token $token
        
        if ($comparison -and $comparison.slug) {
            Test-Endpoint -Method "GET" -Path "/api/v1/comparisons/$($comparison.slug)"
        }
    }
} else {
    Write-Host "${red}⚠️  No token available, skipping protected endpoints${reset}"
}

# 6. Admin endpoints (usando usuario de seed)
Write-Host "`n${yellow}--- Admin Endpoints (using seed user) ---${reset}"
$adminLogin = @{
    email = "test@test.com"
    password = "password"
}
$admin = Test-Endpoint -Method "POST" -Path "/api/v1/auth/login" -Body $adminLogin
if ($admin -and $admin.access_token) {
    Test-Endpoint -Method "GET" -Path "/api/v1/admin/stats" -Token $admin.access_token
}

Write-Host "`n${green}✅ Test completed!${reset}`n"
