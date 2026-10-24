# Nictofilia Go API

Backend REST de Nictofilia, una plataforma social orientada a la creación, descubrimiento, organización y participación en eventos. Este servicio proporciona la infraestructura de API para los clientes que se desarrollen en el futuro; no presupone que exista actualmente una aplicación web o móvil.

## Estado actual

El código implementa:

- Registro de usuarios, perfiles básicos y asignación del rol predeterminado `regular`.
- Inicio de sesión, cierre de sesión y renovación de sesión.
- Access tokens JWT y refresh tokens opacos con rotación; el refresh token se almacena como hash y se entrega mediante una cookie HttpOnly configurable.
- Middleware de autenticación y helper de autorización por rol.
- Endpoints de consulta de usuarios y perfil autenticado.
- API HTTP con Chi, CORS configurable, validación de requests, errores centralizados y logging estructurado.
- Persistencia PostgreSQL, migraciones SQL y tests unitarios.

Las rutas `GET /users/` y `GET /users/{id}` son públicas actualmente. `RequireRole` existe, pero no protege rutas; la autorización basada en permisos no está implementada. Revisa esas políticas antes de exponer el servicio.

Eventos, organizaciones, establecimientos, asistencia, insignias, entradas, pagos, reembolsos y juegos sociales forman parte del contexto futuro del producto y todavía no están implementados.

## Tecnologías

- Go según la directiva `go` de [go.mod](./go.mod).
- Chi para routing y middleware HTTP.
- `database/sql` con PostgreSQL mediante `lib/pq`.
- Migraciones SQL con la herramienta externa `migrate` (`golang-migrate`).
- JWT con `golang-jwt/jwt`, bcrypt mediante `golang.org/x/crypto`.
- Validación con `go-playground/validator`.
- Carga de `.env` con `godotenv`.
- Logging JSON con `log/slog`.

## Arquitectura

El backend está organizado en módulos funcionales: handlers adaptan HTTP, casos de uso coordinan la lógica de aplicación y repositories acceden a los datos. La composición de dependencias es manual desde `cmd/api`; no se impone una Clean Architecture estricta.

Consulta [ARCHITECTURE.md](./ARCHITECTURE.md) para conocer responsabilidades, rutas, flujos y limitaciones actuales.

## Requisitos

- Go compatible con la versión declarada en `go.mod` (actualmente `1.26.2`).
- PostgreSQL.
- GNU Make para usar los targets opcionales del Makefile.
- CLI `migrate` en `PATH` para aplicar migraciones:

  ```powershell
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

## Configuración local

Desde la raíz del proyecto, crea el archivo local de entorno:

```powershell
Copy-Item .env.example .env
```

Configura en `.env` la contraseña de tu usuario PostgreSQL y reemplaza `JWT_SECRET` por un secreto aleatorio privado para este entorno. No compartas ni publiques `.env`; `.env.example` solo contiene placeholders. La aplicación carga `.env` si existe y también acepta variables del entorno.

La base de datos configurada es **`nictofilia`**. Créala manualmente antes de iniciar el servidor. `DATABASE_URL` puede contener un DSN PostgreSQL completo; si está vacío, la aplicación construye la conexión con `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` y `DB_SSLMODE`.

El servidor escucha por defecto en el puerto `8080`. La configuración de JWT usa issuer y audience `nictofilia-go-api`, access token TTL de `15m` y refresh token TTL de `720h`. La cookie refresh se configura mediante las variables `REFRESH_COOKIE_*` (por defecto HttpOnly, `SameSite=lax` y no `Secure` para desarrollo local). CORS permite por defecto `http://localhost:3000`; revisa orígenes, credenciales y cookie al integrar un cliente o desplegar con HTTPS.

## Migraciones

La API no ejecuta migraciones al iniciar. Una vez creada la base `nictofilia`, instala `migrate` y configura para Make las variables PostgreSQL que aparecen en `.env`. Make no carga `.env` por sí mismo. En PowerShell, por ejemplo:

```powershell
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_NAME = "nictofilia"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "<tu-contraseña-local>"
$env:DB_SSLMODE = "disable"
make migrate-up
```

Las migraciones están en `migrations/`. `make migrate-down` y `make migrate-force VERSION=<n>` son operaciones de mantenimiento: verifica cuidadosamente la base destino antes de usarlas. No se ejecutan automáticamente.

## Ejecución

Con PostgreSQL disponible, `.env` configurado y las migraciones aplicadas:

```powershell
go run ./cmd/api
```

También se puede usar `make run`.

## Testing y comprobaciones

```powershell
go test ./...
go build ./...
go vet ./...
```

Los tests actuales son unitarios y no requieren conectarse a PostgreSQL.

## Evolución prevista

Nictofilia busca facilitar el descubrimiento y la participación en eventos públicos y privados, con organizaciones y establecimientos, asistencia, insignias, venta de entradas, pagos, reembolsos y experiencias sociales. Esas capacidades son futuras; no hay actualmente módulos, contratos HTTP ni migraciones para ellas. Se incorporarán progresivamente respetando la arquitectura modular existente.
