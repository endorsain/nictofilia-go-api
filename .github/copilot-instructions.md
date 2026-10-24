# Instrucciones para trabajar en nictofilia-go-api

## Contexto del producto y alcance

- Nictofilia es una red social de eventos. Este repositorio contiene su backend REST.
- Los nuevos módulos de dominio deben incorporarse respetando la arquitectura modular existente y sin mezclar lógica de eventos, organizaciones o pagos con los módulos de autenticación y usuarios.
- Eventos, organizaciones, establecimientos, entradas, pagos y otras funciones de Nictofilia son contexto futuro, no funcionalidades implementadas. No agregues reglas de implementación detalladas hasta que esas tareas se soliciten.
- Consulta [`ARCHITECTURE.md`](../ARCHITECTURE.md) y el código relacionado antes de modificar arquitectura, contratos HTTP, autenticación, persistencia o configuración. El código es la fuente de verdad.

## Arquitectura y diseño

- Conserva las responsabilidades actuales: handlers adaptan HTTP, casos de uso coordinan la lógica de aplicación y repositories acceden a datos.
- Respeta las dependencias entre módulos y la composición explícita desde `cmd/api/main.go`.
- Reutiliza la configuración e infraestructura existentes para PostgreSQL, logging, validación, cookies y errores.
- No impongas Clean Architecture estricta ni agregues interfaces, capas, frameworks o abstracciones sin una necesidad concreta.
- Mantén cambios enfocados y evita dependencias circulares o sistemas paralelos.

## HTTP, seguridad y datos

- Inspecciona los contratos actuales antes de cambiar rutas, métodos, DTOs, nombres JSON, status codes, cookies o autenticación.
- Usa `httptransport.Handle`, `internal/errors` y el middleware HTTP centralizado; no crees envelopes de error alternativos.
- Usa el validator existente para DTOs de request y no expongas errores crudos de validación.
- Mantén DTOs públicos seguros y allowlists explícitas; nunca expongas `PasswordHash`, credenciales, secretos JWT, refresh tokens ni errores internos.
- Conserva bcrypt, la validación configurada del JWT y el almacenamiento de refresh tokens como hashes con rotación transaccional.
- Mantén SQL parametrizado y transacciones cuando la operación requiera atomicidad.
- Propaga `context.Context` a operaciones de aplicación y persistencia.
- No cambies silenciosamente la semántica de autenticación, autorización ni respuestas exitosas.

## Calidad, documentación y convenciones

- Inspecciona el código existente antes de editar y ejecuta los tests relevantes; mantén o amplía tests cuando corresponda.
- Sigue las convenciones de Go: identificadores y paquetes en inglés; comentarios técnicos en español cuando aporten contexto.
- Actualiza `ARCHITECTURE.md` cuando cambien decisiones arquitectónicas o contratos relevantes.
- Usa `gofmt`; no actualices dependencias no relacionadas ni introduzcas herramientas innecesarias.

## Git y operaciones

- No realices commits ni pushes automáticamente ni ejecutes operaciones Git destructivas.
- No ejecutes migraciones ni operaciones contra bases de datos sin autorización explícita.
- No incluyas `.env`, secretos ni credenciales en cambios o documentación.
