# PassGen CLI

CLI personal para generar y almacenar contraseñas seguras, cifradas localmente con AES-256-GCM.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Plataforma](https://img.shields.io/badge/Plataforma-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

---

## Características

- Generación criptográficamente segura usando `crypto/rand`
- Cifrado AES-256-GCM con derivación de clave Argon2id
- Vault cifrado en `~/.passgen/vault.enc` — solo tú puedes leerlo
- Copia automática de la contraseña al portapapeles
- Subcomando `list` para ver todas tus entradas en tabla
- Sin dependencias externas en tiempo de ejecución — un solo binario

---

## Requisitos

- [Go 1.21+](https://go.dev/dl/)
- Git (para clonar el repositorio)

---

## Instalación

### 1. Clonar y compilar

```bash
git clone https://github.com/VictorRoe/passgen-cli.git
cd passgen-cli
```

### 2. Instalar el binario según tu sistema operativo

#### Windows (PowerShell o Git Bash)

```bash
go build -o "$GOPATH/bin/passgen.exe" .
```

> `go install` crearía `passgen-cli.exe` (por el nombre del módulo). Con `-o` forzamos el nombre a `passgen`.

Verificar que `%GOPATH%\bin` está en el PATH:

```powershell
$env:Path -split ';' | Select-String "go\\bin"
```

Si no aparece, agrégalo (una sola vez) y reabre PowerShell:

```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\go\bin", "User")
```

#### macOS / Linux

```bash
go build -o "$GOPATH/bin/passgen" .
```

Verificar que `$GOPATH/bin` está en el PATH (agrega al `.bashrc` o `.zshrc` si no está):

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 3. Verificar la instalación

```bash
passgen --help
```

---

## Primer uso

La primera vez que ejecutes `passgen`, te pedirá una **Master Password**. Esa contraseña protege todo tu vault — elige una que no olvides, no hay forma de recuperarla.

```bash
passgen email@gmail.com github -C 24 -N -S
```

```
✅ Password copiado al portapapeles.
Master Password:
🔐 Registro cifrado y guardado con éxito.
```

La contraseña generada queda en tu portapapeles lista para pegar.

---

## Comandos

### Generar y guardar una contraseña

```bash
passgen <email> <servicio> [flags]
```

**Ejemplos:**

```bash
# Contraseña de 16 caracteres (por defecto)
passgen user@gmail.com netflix

# 24 caracteres con números y símbolos
passgen user@gmail.com github -C 24 -N -S

# Solo números, longitud 20
passgen user@gmail.com banco -C 20 -N
```

**Flags:**

| Flag | Descripción | Por defecto |
|------|-------------|-------------|
| `-C, --chars` | Longitud de la contraseña (mínimo 8) | `16` |
| `-N, --numbers` | Incluir números `0-9` | `false` |
| `-S, --symbols` | Incluir símbolos `@#$%&*()!` | `false` |

---

### Listar contraseñas guardadas

```bash
passgen list
```

```
Master Password:
EMAIL                            SERVICIO             PASSWORD                         CREADO
--------------------------------------------------------------------------------------------------------
victorroe@gmail.com              github               aB3$kP9!mXq2Lz@n                2026-03-16 14:32
victorroe@gmail.com              netflix              mNp7Wd4Kqr1Ytz8V                2026-03-16 14:45
```

---

## Seguridad

| Componente | Detalle |
|------------|---------|
| Generación | `crypto/rand` — aleatoriedad criptográfica |
| Cifrado | AES-256-GCM (cifrado autenticado) |
| Derivación de clave | Argon2id — `time=3`, `memory=64MB` |
| Vault | `~/.passgen/vault.enc` — permisos `0600` |
| Directorio | `~/.passgen/` — permisos `0700` |
| Escritura | Atómica vía archivo `.tmp` + rename |
| Memoria | Master password y clave zeroeados tras uso |

El vault no puede descifrarse sin la Master Password. Si la pierdes, los datos no son recuperables.

---

## Cómo funciona internamente

```
passgen <email> <servicio> [flags]
        │
        ├─ Genera password con crypto/rand
        ├─ Copia al portapapeles
        ├─ Pide Master Password (sin eco en terminal)
        ├─ Deriva clave AES-256 con Argon2id + salt aleatorio
        ├─ Descifra el vault existente (si hay entradas previas)
        ├─ Agrega la nueva entrada en memoria
        └─ Re-cifra y escribe el vault atómicamente
```

---

## Actualizar el binario

Tras modificar el código, vuelve a compilar:

```bash
# Windows
go build -o "$GOPATH/bin/passgen.exe" .

# macOS / Linux
go build -o "$GOPATH/bin/passgen" .
```
