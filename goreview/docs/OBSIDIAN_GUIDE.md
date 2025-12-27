# Guia de Integracion con Obsidian

Esta guia te ayudara a configurar y usar la integracion de GoReview con Obsidian, incluso si nunca has usado Obsidian antes.

---

## Tabla de Contenidos

1. [Que es Obsidian?](#que-es-obsidian)
2. [Instalacion de Obsidian](#instalacion-de-obsidian)
3. [Crear tu Primer Vault](#crear-tu-primer-vault)
4. [Configurar GoReview](#configurar-goreview)
5. [Exportar Reviews a Obsidian](#exportar-reviews-a-obsidian)
6. [Entendiendo las Notas Generadas](#entendiendo-las-notas-generadas)
7. [Funcionalidades Avanzadas](#funcionalidades-avanzadas)
8. [Preguntas Frecuentes](#preguntas-frecuentes)

---

## Que es Obsidian?

Obsidian es una aplicacion gratuita para tomar notas que funciona con archivos Markdown locales en tu computadora. A diferencia de otras apps como Notion o Evernote, tus notas:

- **Son archivos locales**: No dependen de ningun servidor
- **Estan en formato Markdown**: Texto plano que puedes leer en cualquier editor
- **Se conectan entre si**: Puedes crear links entre notas como una wiki personal

### Por que usar Obsidian con GoReview?

- **Historial de reviews**: Todas tus revisiones de codigo quedan documentadas
- **Busqueda potente**: Encuentra patrones, errores recurrentes, soluciones pasadas
- **Conexiones**: Ve como los issues se relacionan entre proyectos
- **Offline**: Todo funciona sin internet

---

## Instalacion de Obsidian

### Windows

1. Ve a [obsidian.md/download](https://obsidian.md/download)
2. Descarga el instalador para Windows
3. Ejecuta el archivo `.exe` descargado
4. Sigue el asistente de instalacion

### macOS

1. Ve a [obsidian.md/download](https://obsidian.md/download)
2. Descarga el archivo `.dmg`
3. Abre el `.dmg` y arrastra Obsidian a tu carpeta Aplicaciones

### Linux

```bash
# Usando Snap
sudo snap install obsidian --classic

# O descarga el AppImage desde obsidian.md/download
```

---

## Crear tu Primer Vault

Un **vault** es simplemente una carpeta donde Obsidian guardara tus notas. Puedes tener multiples vaults para diferentes propositos.

### Paso 1: Abrir Obsidian

Al abrir Obsidian por primera vez, veras una pantalla de bienvenida.

### Paso 2: Crear un Vault

1. Haz clic en **"Create new vault"** (Crear nuevo vault)
2. Escribe un nombre, por ejemplo: `MisNotas` o `Desarrollo`
3. Elige una ubicacion:
   - Windows: `C:\Users\TuUsuario\Documents\ObsidianVaults\`
   - macOS: `~/Documents/ObsidianVaults/`
   - Linux: `~/Documents/ObsidianVaults/`
4. Haz clic en **"Create"**

### Paso 3: Explorar la Interfaz

```
+------------------------------------------+
|  [<] [>]  MisNotas          [-] [O] [X] |
+----------+-------------------------------+
|          |                               |
| ARCHIVOS |      AREA DE EDICION          |
|          |                               |
| > GoReview                               |
|   - review-001.md                        |
|   - review-002.md                        |
|          |                               |
+----------+-------------------------------+
```

- **Panel izquierdo**: Lista de archivos y carpetas
- **Panel central**: Editor de notas
- **Panel derecho** (opcional): Vista previa, links, etc.

### Paso 4: Obtener la Ruta del Vault

Necesitaras la ruta completa para configurar GoReview:

**Windows:**
```
C:\Users\TuUsuario\Documents\ObsidianVaults\MisNotas
```

**macOS/Linux:**
```
/Users/TuUsuario/Documents/ObsidianVaults/MisNotas
```

> **Tip**: En Obsidian, haz clic en el icono de engranaje (Configuracion) > "About" para ver la ruta del vault actual.

---

## Configurar GoReview

Hay dos formas de configurar la exportacion a Obsidian:

### Opcion 1: Archivo de Configuracion (Recomendado)

Crea o edita el archivo `.goreview.yaml` en la raiz de tu proyecto:

```yaml
# .goreview.yaml

# ... otras configuraciones ...

export:
  obsidian:
    # Activa la exportacion automatica despues de cada review
    enabled: true

    # Ruta a tu vault de Obsidian (OBLIGATORIO)
    # Windows: usa barras normales o dobles
    vault_path: "C:/Users/TuUsuario/Documents/ObsidianVaults/MisNotas"
    # macOS/Linux:
    # vault_path: "~/Documents/ObsidianVaults/MisNotas"

    # Carpeta dentro del vault donde se guardaran los reviews
    folder_name: "GoReview"

    # Incluir tags de Obsidian (#security, #bug, etc.)
    include_tags: true

    # Usar callouts de Obsidian para resaltar issues
    include_callouts: true

    # Crear links a reviews anteriores del mismo proyecto
    include_links: true
    link_to_previous: true

    # Tags adicionales para todas las notas
    custom_tags:
      - "code-review"
      # - "mi-equipo"
      # - "proyecto-x"
```

### Opcion 2: Flags en Linea de Comandos

Si prefieres no usar archivo de configuracion:

```bash
goreview review --staged --export-obsidian --obsidian-vault "C:/Users/TuUsuario/Documents/ObsidianVaults/MisNotas"
```

---

## Exportar Reviews a Obsidian

### Metodo 1: Automatico con Review

Ejecuta un review y exporta automaticamente:

```bash
# Si tienes enabled: true en config
goreview review --staged

# O forzar exportacion con flag
goreview review --staged --export-obsidian
```

### Metodo 2: Exportar desde JSON

Primero genera un reporte JSON, luego exportalo:

```bash
# Paso 1: Generar JSON
goreview review --staged --format json -o report.json

# Paso 2: Exportar a Obsidian
goreview export obsidian --from report.json --vault ~/MisNotas
```

### Metodo 3: Pipeline con Stdin

```bash
# Todo en un comando
goreview review --staged --format json | goreview export obsidian --vault ~/MisNotas
```

### Verificar la Exportacion

Despues de exportar, veras un mensaje como:

```
Exported to Obsidian: C:\Users\...\MisNotas\GoReview\mi-proyecto\review-001-2024-12-27.md
```

Abre Obsidian y navega a la carpeta `GoReview` para ver tu nota.

---

## Entendiendo las Notas Generadas

Cada review genera una nota con esta estructura:

### Frontmatter (Metadatos)

En la parte superior, entre `---`, hay metadatos en formato YAML:

```yaml
---
date: 2024-12-27T10:30:00Z
project: mi-proyecto
branch: feature/nueva-funcion
commit: abc123d
files_reviewed: 5
total_issues: 3
severity:
  critical: 0
  error: 1
  warning: 2
  info: 0
average_score: 85
tags:
  - goreview
  - code-review
  - security
---
```

Estos metadatos permiten:
- Buscar notas por fecha, proyecto, severidad
- Filtrar en el Graph View de Obsidian
- Crear queries con plugins como Dataview

### Tags de Obsidian

Debajo del titulo veras tags como:

```
#goreview #code-review #security #bug
```

En Obsidian, puedes hacer clic en cualquier tag para ver todas las notas con ese tag.

### Tabla de Resumen

```markdown
| Metric | Value |
|--------|-------|
| Files Reviewed | 5 |
| Total Issues | 3 |
| Average Score | 85/100 |
```

### Issues con Callouts

Los issues se muestran con callouts de Obsidian que tienen colores segun severidad:

```markdown
> [!danger] :red_circle: **[security]** SQL injection vulnerability
> **Location:** Line 45
> **Suggestion:** Use parameterized queries
```

Los tipos de callout son:
- `[!danger]` - Rojo - Issues criticos
- `[!warning]` - Naranja - Errores
- `[!caution]` - Amarillo - Advertencias
- `[!info]` - Azul - Informacion

### Links a Reviews Anteriores

Al final de cada nota, veras links a reviews anteriores:

```markdown
## Related Reviews

- [[review-001-2024-12-25]]
- [[review-002-2024-12-26]]
```

Haz clic en cualquier link para navegar a esa nota.

---

## Funcionalidades Avanzadas

### Buscar en tus Reviews

En Obsidian, presiona `Ctrl+Shift+F` (Windows/Linux) o `Cmd+Shift+F` (macOS) para buscar:

- `tag:#security` - Todos los reviews con issues de seguridad
- `tag:#critical` - Reviews con issues criticos
- `SQL injection` - Buscar texto especifico
- `path:GoReview/mi-proyecto` - Solo en un proyecto

### Graph View (Vista de Grafo)

Obsidian puede mostrar tus notas como un grafo conectado:

1. Presiona `Ctrl+G` o haz clic en el icono de grafo
2. Veras tus reviews conectados entre si
3. Los colores pueden indicar proyectos o severidad

### Usar con Plugin Dataview

Si instalas el plugin **Dataview**, puedes crear queries:

```dataview
TABLE date, total_issues, average_score
FROM "GoReview"
WHERE total_issues > 0
SORT date DESC
LIMIT 10
```

Esto mostrara una tabla con tus ultimos 10 reviews que tuvieron issues.

### Templates Personalizados

Puedes crear tu propio template. Crea un archivo y configuralo:

```yaml
# .goreview.yaml
export:
  obsidian:
    template_file: "./mi-template.md"
```

El template usa sintaxis de Go templates. Consulta `internal/export/obsidian_template.go` como referencia.

---

## Preguntas Frecuentes

### No encuentro mis notas en Obsidian

1. Verifica que la ruta del vault sea correcta
2. En Obsidian, haz clic derecho en el panel izquierdo > "Reveal in Finder/Explorer"
3. Navega a la carpeta `GoReview`

### El export falla con "vault not found"

- Verifica que la ruta existe
- En Windows, usa `/` o `\\` en las rutas
- Asegurate de que Obsidian haya creado la carpeta `.obsidian` dentro del vault

### Como cambio el nombre de la carpeta?

```yaml
export:
  obsidian:
    folder_name: "CodeReviews"  # Cambiar de "GoReview" a otro nombre
```

### Puedo exportar a multiples vaults?

Actualmente solo se puede exportar a un vault por configuracion. Para exportar a otro vault, usa el flag:

```bash
goreview export obsidian --from report.json --vault ~/OtroVault
```

### Los callouts no se ven con colores

Los callouts son una caracteristica nativa de Obsidian desde la version 0.14. Si no los ves:

1. Actualiza Obsidian a la ultima version
2. Asegurate de estar en modo "Reading View" o "Live Preview"

### Como desactivo la exportacion automatica?

```yaml
export:
  obsidian:
    enabled: false  # Cambiar a false
```

O simplemente no uses el flag `--export-obsidian`.

### Puedo sincronizar mi vault con la nube?

Si, puedes poner tu vault en:
- **Dropbox, Google Drive, OneDrive**: Sincronizacion automatica
- **Git**: Versionar tus notas (hay plugins para esto)
- **Obsidian Sync**: Servicio de pago oficial

---

## Resumen de Comandos

```bash
# Review con exportacion automatica (si enabled: true)
goreview review --staged

# Forzar exportacion
goreview review --staged --export-obsidian

# Especificar vault
goreview review --staged --export-obsidian --obsidian-vault "~/MiVault"

# Exportar desde archivo JSON
goreview export obsidian --from report.json --vault "~/MiVault"

# Exportar con tags adicionales
goreview export obsidian --from report.json --vault "~/MiVault" --tags sprint-42,backend

# Exportar sin callouts
goreview export obsidian --from report.json --vault "~/MiVault" --no-callouts

# Pipeline completo
goreview review --staged -f json | goreview export obsidian --vault "~/MiVault"
```

---

## Soporte

Si tienes problemas:

1. Verifica la configuracion con `goreview config show`
2. Revisa que la ruta del vault sea accesible
3. Abre un issue en el repositorio de GoReview
