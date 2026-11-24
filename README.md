# PassGen CLI

CLI personal para generar y almacenar contraseñas seguras, cifradas localmente con AES-256-GCM.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Plataforma](https://img.shields.io/badge/Plataforma-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

---

## Características

- Generación criptográficamente segura usando `crypto/rand`
- Cifrado AES-256-GCM con derivación de clave Argon2id
- Vault cifrado en `~/.passgen/vault.enc` — solo tú puedes leerlo
- Copia automática de la contraseña al portapapeles
- Subcomando `list` para ver todas tus entradas en tabla
- Subcomando `delete` interactivo (o por ID) con confirmación de una tecla
- Subcomando `copy` interactivo (o por ID) que borra el portapapeles a los 30 segundos
- Un solo binario, sin runtime (en Linux el portapapeles necesita `xclip`, `xsel` o `wl-clipboard`)

---

## Requisitos

- [Go 1.25+](https://go.dev/dl/) (lo exige `go.mod`)
- Git (para clonar el repositorio)
- Solo Linux: `xclip`, `xsel` o `wl-clipboard` para el portapapeles (p. ej. `sudo apt install xclip`)

---

## Instalación

### 1. Clonar

```bash
git clone https://github.com/VictorRoe/passgen-cli.git
cd passgen-cli
```

### 2. Compilar e instalar el binario

El binario se instala en la carpeta `bin` de tu `GOPATH`. Usamos `go env GOPATH` en vez de `$GOPATH` porque esa variable normalmente no está definida (Go usa un valor por defecto, p. ej. `C:\Users\<tú>\go` o `~/go`).

#### Windows (PowerShell o Git Bash)

```bash
go build -o "$(go env GOPATH)/bin/passgen.exe" .
```

> `go install` crearía `passgen-cli.exe` (por el nombre del módulo). Con `-o` forzamos el nombre a `passgen`.

Verificar que esa carpeta está en el PATH (PowerShell):

```powershell
$bin = "$(go env GOPATH)\bin"
$env:Path -split ';' | Where-Object { $_ -eq $bin }
```

Si no aparece nada, agrégala al PATH de tu usuario (una sola vez) y reabre la terminal:

```powershell
$bin = "$(go env GOPATH)\bin"
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
[Environment]::SetEnvironmentVariable("Path", "$userPath;$bin", "User")
```

#### macOS / Linux

```bash
go build -o "$(go env GOPATH)/bin/passgen" .
```

Verificar que esa carpeta está en el PATH (agrega la línea a tu `.bashrc` o `.zshrc` si no está):

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
ID        EMAIL                            SERVICIO             PASSWORD                         CREADO
-----------------------------------------------------------------------------------------------------------------
3f9a1c2e  user@gmail.com                   github               aB3$kP9!mXq2Lz@n                 2026-03-16 14:32
8b7d4e10  user@gmail.com                   netflix              mNp7Wd4Kqr1Ytz8V                 2026-03-16 14:45
```

---

### Eliminar una contraseña

**Modo interactivo** — elige la entrada de una lista:

```bash
passgen delete
```

```
Selecciona la entrada a eliminar (↑/↓, Enter, Esc/Ctrl+C para salir):
> 3f9a1c2e  github                (user@gmail.com)
  8b7d4e10  netflix               (user@gmail.com)
```

**Por ID** — usa los primeros 8+ caracteres del ID que muestra `passgen list`:

```bash
passgen delete 3f9a1c2e
```

En ambos modos se pide confirmación:

```
¿Eliminar 'github' (user@gmail.com)? [Enter] confirmar · [Esc/Ctrl+C] cancelar
```

| Tecla | Lista | Confirmación |
|-------|-------|--------------|
| `↑`/`↓` o `k`/`j` | Mover la selección | — |
| `Enter` | Elegir entrada | Eliminar |
| `Esc` / `Ctrl+C` | Salir sin borrar | Cancelar sin borrar |

Cualquier otra tecla en la confirmación se ignora.

---

### Copiar una contraseña

**Modo interactivo** — elige la entrada de una lista:

```bash
passgen copy
```

**Por ID** — usa los primeros 8+ caracteres del ID que muestra `passgen list`:

```bash
passgen copy 3f9a1c2e
```

La contraseña se copia al portapapeles (nunca se muestra en pantalla) y empieza una cuenta atrás:

```
✅ Password de 'github' copiado. Se borrará en 30s · [Enter/Esc/Ctrl+C] borrar ahora
🧹 Portapapeles borrado.
```

- A los 30 segundos, o al pulsar `Enter`, `Esc` o `Ctrl+C`, el portapapeles se borra.
- Si mientras tanto copiaste otra cosa, no se toca: solo se borra si aún contiene esa contraseña.
- En la lista, `Esc` o `Ctrl+C` salen sin copiar nada.

> ⚠️ Si tienes activado el **historial del portapapeles de Windows** (`Win+V`) o la sincronización entre dispositivos, la contraseña puede quedar guardada en ese historial aunque `passgen` borre el portapapeles. Bórrala desde `Win+V` o desactiva el historial en *Configuración → Sistema → Portapapeles*.

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

Tras modificar el código (o hacer `git pull`), vuelve a compilar desde la raíz del repo:

```bash
# Windows (PowerShell o Git Bash)
go build -o "$(go env GOPATH)/bin/passgen.exe" .

# macOS / Linux
go build -o "$(go env GOPATH)/bin/passgen" .
```

Para comprobar desde qué commit está compilado el binario:

```bash
go version -m "$(go env GOPATH)/bin/passgen.exe"   # busca vcs.revision
```
