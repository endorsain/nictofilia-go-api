# Arquitectura de nictofilia-go-api

## Propósito

`nictofilia-go-api` es el backend REST de Nictofilia, una plataforma social orientada a la creación, descubrimiento, organización y participación en eventos. El backend expone servicios HTTP y coordina la persistencia y las reglas de aplicación para los clientes frontend o móviles que puedan desarrollarse; esos clientes no forman parte de este repositorio actualmente.

## Identidad y alcance actual

- **Proyecto y aplicación:** `nictofilia-go-api`.
- **Producto:** Nictofilia, plataforma social de eventos.
- **Lenguaje:** Go.
- **Persistencia:** PostgreSQL mediante `database/sql` y `lib/pq`.
- **Diseño:** módulos funcionales, separación de handlers, casos de uso y repositories, e inyección manual de dependencias. Comparte ideas de Ports and Adapters donde hay límites concretos; no pretende imponer Clean Architecture estricta.

### Funcionalidades implementadas

- Usuarios, perfiles básicos y roles `admin`, `moderator` y `regular`; el registro asigna el rol `regular`.
- Registro, login, logout y refresh de sesión.
- Access tokens JWT firmados con HS256, validados comprobando algoritmo, issuer, audience y expiración.
- Refresh tokens opacos generados criptográficamente; se persiste su hash SHA-256 y la rotación se realiza transaccionalmente.
- Cookie configurable para refresh, incluyendo `HttpOnly`, `Secure`, `SameSite`, dominio y path.
- Middleware de autenticación y helper `RequireRole`.
- API HTTP, routing, CORS, validación de requests, errores HTTP centralizados, logging JSON y conexión PostgreSQL compartida.
- Migraciones SQL de infraestructura y tests unitarios.

### Funcionalidades previstas

El contexto de producto contempla eventos públicos y privados, organizaciones, establecimientos, asistencia, insignias, venta y gestión de entradas, pagos, reembolsos, juegos sociales y otras experiencias para participantes. Estas funcionalidades **no están implementadas**: no hay módulos, contratos HTTP ni esquemas de base de datos definidos para ellas.

## Estructura y responsabilidades

- `cmd/api/`: composition root y ejecutable. Carga configuración, inicializa logging y PostgreSQL, construye módulos, registra rutas y gestiona el cierre del servidor.
- `internal/auth/`: registro, login, logout, refresh, servicio JWT, persistencia de refresh tokens y middleware de autenticación/autorización.
- `internal/users/`: modelos de usuarios y perfiles, DTOs de respuesta, casos de uso de consulta y adaptador PostgreSQL.
- `internal/config/`: lectura de configuración desde variables de entorno y `.env`; agrupa aplicación, base de datos, servidor, CORS, autenticación, cookies y logging.
- `internal/infrastructure/postgres/`: conexión PostgreSQL compartida, ping de inicio y configuración del pool.
- `internal/transport/http/`: servidor y router Chi, CORS, cookies, decode JSON, validación, helpers HTTP y middleware de errores.
- `internal/errors/`: errores de aplicación y representación HTTP pública común.
- `internal/logging/`: logger JSON basado en `log/slog`.
- `migrations/`: migraciones SQL versionadas en pares `up` y `down`.

La composición de dependencias es explícita desde `cmd/api/main.go`. Los handlers adaptan HTTP, los casos de uso coordinan la lógica de aplicación y los repositories ejecutan SQL parametrizado. El módulo `auth` depende de `users` para el registro y el usuario requerido por login; el composition root conecta el repository de usuarios con autenticación. No se agrega una capa global adicional ni se requiere crear interfaces donde no haya un límite concreto.

## Flujos HTTP y contratos actuales

Las rutas y respuestas existentes son el contrato actual; deben preservarse salvo que un cambio explícito lo requiera.

| Método | Ruta | Acceso actual | Comportamiento |
|---|---|---|---|
| `POST` | `/auth/register` | Público | Crea usuario, perfil y rol predeterminado; responde `201` con `id`, `username` y `email`. |
| `POST` | `/auth/login` | Público | Valida credenciales, devuelve `access_token` y establece la cookie de refresh. |
| `POST` | `/auth/refresh` | Cookie de refresh válida | Rota el refresh token, renueva la cookie y devuelve `access_token`. |
| `POST` | `/auth/logout` | Público | Revoca el refresh token si está presente, limpia la cookie y responde `204`. |
| `GET` | `/users/me` | Access token | Devuelve el usuario asociado al subject del token y su perfil. |
| `GET` | `/users/` | Público actualmente | Devuelve resúmenes de usuarios. |
| `GET` | `/users/{id}` | Público actualmente | Devuelve usuario y perfil mediante campos permitidos explícitamente. |

### Registro y usuarios

Los handlers decodifican y validan los DTOs de registro y login con `go-playground/validator`. El caso de uso de registro genera el hash bcrypt. El repository crea usuario, perfil y rol `regular` dentro de una transacción. Las respuestas usan DTOs seguros y no incluyen `PasswordHash`.

### Autenticación y sesiones

El login compara la contraseña con bcrypt, consulta los roles y emite un JWT HS256. La validación fija el algoritmo y verifica issuer, audience y expiración. El access token se devuelve en JSON y se envía como `Authorization: Bearer` en rutas protegidas.

El refresh token es opaco y aleatorio; solo se guarda su hash. El refresh verifica estado y expiración, carga roles vigentes y rota el token con revocación y creación atómicas. Logout revoca el token si hay cookie y la elimina; una cookie ausente no impide responder `204`.

El middleware de autenticación agrega subject y roles al contexto de request. `RequireRole` comprueba un rol por nombre, pero no está conectado actualmente a una ruta. Existen tablas de permisos y relaciones, pero no se implementa autorización basada en permisos.

### HTTP, validación y errores

El router Chi incorpora middleware común y CORS configurable. `RequestValidator` se inyecta en registro y login. El adaptador HTTP central representa errores con `{"error":{"code":"...","message":"..."}}`; evita exponer errores internos o de validación. Los errores inesperados 5xx se registran con request ID, método y ruta; los middleware de autenticación y autorización usan el mismo formato público.

## Persistencia

El cliente PostgreSQL compartido configura un pool, ejecuta `PingContext` al iniciar y cierra la conexión durante el apagado. Los repositories de `auth` y `users` propagan `context.Context` y usan consultas parametrizadas.

Las migraciones actuales habilitan `pgcrypto` y crean `users`, `user_profiles`, `refresh_tokens`, `roles`, `permissions`, `user_roles` y `role_permissions`. Seeds incorporan los roles generales `admin`, `moderator` y `regular`; no hay permisos de dominio precargados. El nombre configurado de la base de datos de Nictofilia es `nictofilia`. Las migraciones no se ejecutan automáticamente al iniciar el backend.

La creación de usuario, perfil y asignación del rol predeterminado es transaccional. La rotación de refresh tokens también agrupa revocación y creación en una transacción.

## Configuración, cookies, CORS y logging

`internal/config.Load` carga opcionalmente `.env` mediante `godotenv` y centraliza `APP_*`, `DATABASE_URL`/`DB_*`, `SERVER_*`, `CORS_*`, `JWT_*`, `REFRESH_*` y `LOG_LEVEL`. `JWT_SECRET` debe estar configurado para inicializar autenticación; no tiene un secreto por defecto en el código.

Los valores por defecto son para desarrollo local: aplicación `nictofilia-go-api`, base `nictofilia`, puerto HTTP `8080`, issuer y audience `nictofilia-go-api`, access token TTL `15m`, refresh token TTL `720h` y CORS para `http://localhost:3000`. Si `DATABASE_URL` está vacío, el DSN se construye desde `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` y `DB_SSLMODE`.

La cookie refresh usa por defecto nombre `refresh_token`, path `/auth`, `HttpOnly=true`, `Secure=false` y `SameSite=lax`. En un despliegue HTTPS deben revisarse `Secure`, `SameSite`, dominio y orígenes CORS de acuerdo con el cliente. El logger usa salida JSON y nivel configurable; no registra el secreto JWT ni la contraseña de PostgreSQL al iniciar.

## Evolución de módulos Nictofilia

Los módulos específicos del producto se incorporarán progresivamente bajo `internal/` y se compondrán explícitamente desde `cmd/api/main.go`. Cada área mantendrá los handlers como adaptadores HTTP, los casos de uso para la coordinación de aplicación y los repositories para persistencia. Los futuros módulos de eventos, organizaciones, establecimientos y pagos no deben mezclar su lógica con autenticación ni usuarios.

Los contratos HTTP, modelos y migraciones de esas áreas se definirán cuando se implementen; este documento no anticipa su estructura.

## Limitaciones conocidas

- `GET /users/` y `GET /users/{id}` son públicos en el router actual; la política debe revisarse antes de exponer la API.
- `RequireRole` existe, pero no protege rutas. La autorización basada en permisos no está implementada aunque existan tablas de permisos.
- La validación compartida está conectada a registro y login; las rutas de consulta actuales no reciben DTOs JSON.
- `UserRepository` conserva métodos que no se usan en los flujos compuestos actuales.
- Los tests existentes cubren validación, errores HTTP, middleware y respuestas seguras, pero no repositories PostgreSQL, transacciones ni todos los casos de uso.
- Los errores al cargar `.env` y los valores inválidos de durations o booleanos usan silenciosamente los defaults existentes.
- La CLI `migrate` debe instalarse por separado; el Makefile no instala herramientas.

## Origen del código

La estructura inicial se derivó de `go-api-starter`, en el repositorio `projects-templates`. Este repositorio ya es el backend concreto de Nictofilia y no depende del starter en tiempo de ejecución.
