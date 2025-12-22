# GoReview CLI

Herramienta de linea de comandos para code review con IA.

## Compilar

```bash
make build
```

## Ejecutar

```bash
./build/goreview review
```

## Tests

```bash
make test
```

## Comandos

- `review`: Analiza codigo y genera feedback
- `commit`: Genera mensaje de commit con IA
- `doc`: Genera documentacion de cambios
- `config`: Muestra configuracion actual
- `init-project`: Inicializa proyecto con config
- `version`: Muestra version

## Flags globales

- `--config, -c`: Archivo de configuracion
- `--verbose, -v`: Output detallado
- `--quiet, -q`: Solo errores
