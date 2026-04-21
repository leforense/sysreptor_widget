# Reptor Widget — XFCE Panel + ksnip Addon for SysReptor

> **[English below / Inglês abaixo]**

---

## Português (pt-br)

### O que é isso?

Dois utilitários para integrar o [SysReptor](https://github.com/Syslifters/sysreptor) ao desktop Kali Linux (XFCE):

| Ferramenta | Binário | Função |
|---|---|---|
| **Panel Widget** | `reptor_widget` | Mostra o projeto ativo no xfce4-panel; clique para trocar de projeto |
| **ksnip Addon** | `reptor-upload` | Hook do ksnip — ao salvar um print, faz upload direto para um Finding no SysReptor |

**Por que existe?** Durante um pentest com vários projetos abertos, basta bater o olho no painel para saber em qual projeto você está, e trocar com um único clique. O addon do ksnip elimina a etapa manual de abrir o browser para anexar evidências.

---

### Pré-requisitos

- Kali Linux com XFCE
- Go 1.21+
- `zenity` (já incluso no Kali)
- `xfce4-genmon-plugin` (para o widget do painel)
- `ksnip` (para o addon de upload)
- Acesso a uma instância do SysReptor

---

### Instalação

#### 1. Configuração de Autorização

O widget lê `server`, `token` e `project_id` do arquivo padrão do sysreptor CLI:

```
~/.sysreptor/config.yaml
```

Se você usa o sysreptor CLI (`reptor`), esse arquivo já existe. Para verificar:

```bash
cat ~/.sysreptor/config.yaml
```

Exemplo de conteúdo esperado:
```yaml
server: http://reptor-local:8080
token: seu_api_token_aqui
project_id: uuid-do-projeto-ativo
```

**Se você NÃO usa o sysreptor CLI**, descomente as linhas `server` e `token` no arquivo de configuração do widget (veja seção abaixo).

O token de API pode ser gerado em: **SysReptor → Usuário → API Tokens**

---

#### 2. Arquivo de Configuração do Widget

Copie o exemplo e edite:

```bash
mkdir -p ~/.config/reptor-widget
cp config.conf.example ~/.config/reptor-widget/config.conf
nano ~/.config/reptor-widget/config.conf
```

Campos disponíveis:

| Campo | Descrição | Padrão |
|---|---|---|
| `language` | Idioma dos diálogos: `pt-br` ou `en-us` | `en-us` |
| `template_uuid` | UUID do template usado ao criar novos findings | — |
| `api_upload_field` | Nome do campo Markdown no Design do projeto | `api_upload` |
| `server` | *(opcional)* URL do SysReptor — use se não tiver sysreptor CLI | — |
| `token` | *(opcional)* API Token — use se não tiver sysreptor CLI | — |

**Como obter o `template_uuid`:**
1. Abra o SysReptor no browser
2. Vá em **Findings → Templates** e escolha seu template
3. O UUID está na URL:
   ```
   http://reptor-local:8080/templates/7b0dd839-259a-4fb1-ad08-179ec2514a95/
                                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                      este é o template_uuid
   ```

**O que é `api_upload_field`:**  
Um campo do tipo **Markdown** no Design do seu projeto SysReptor. É onde as imagens enviadas pelo ksnip serão acumuladas. Crie um campo interno (não publicado no PDF final) no seu Design — por exemplo `api_upload`. O nome que você definir aqui deve ser idêntico ao nome do campo no Design.

---

#### 3. Compilar e Instalar

**Panel Widget:**
```bash
go build -o reptor_widget main.go
sudo cp reptor_widget /usr/local/bin/
```

**ksnip Addon:**
```bash
cd ksnip_addon_reptor
go build -o reptor-upload ksnip-reptor.go
sudo cp reptor-upload /usr/local/bin/
```

---

#### 4. Configurar o Panel Widget no XFCE

1. Clique direito no painel → **Add New Items** → **Generic Monitor (GenMon)**
2. Clique direito no GenMon adicionado → **Properties**
3. Configure:
   - **Command:** `/usr/local/bin/reptor_widget`
   - **Period:** `30` (segundos entre atualizações)
   - Marque **Run in terminal:** não
4. O painel vai exibir o nome do projeto ativo em verde. Clique para abrir o seletor.

---

#### 5. Configurar o ksnip

1. Abra o ksnip → **Options → Settings → Script**
2. Em **Post-capture script**, insira:
   ```
   /usr/local/bin/reptor-upload %1
   ```
   O `%1` é substituído pelo caminho do arquivo salvo automaticamente pelo ksnip.

---

### Uso

**Trocar de projeto:** clique no widget no painel → selecione o projeto na lista.

**Enviar evidência:** capture um print com ksnip → ao salvar, uma janela aparece com a lista de findings do projeto ativo → selecione um finding existente ou crie um novo → a imagem é enviada e o campo Markdown é atualizado.

---

---

## English (en-us)

### What is this?

Two utilities to integrate [SysReptor](https://github.com/Syslifters/sysreptor) into the Kali Linux XFCE desktop:

| Tool | Binary | Purpose |
|---|---|---|
| **Panel Widget** | `reptor_widget` | Shows the active project in the xfce4-panel; click to switch projects |
| **ksnip Addon** | `reptor-upload` | ksnip hook — when saving a screenshot, uploads it directly to a Finding in SysReptor |

**Why does this exist?** When running multiple pentests simultaneously, a glance at the panel tells you which project is active, and a single click switches it. The ksnip addon removes the manual step of opening a browser to attach evidence.

---

### Prerequisites

- Kali Linux with XFCE
- Go 1.21+
- `zenity` (included in Kali)
- `xfce4-genmon-plugin` (for the panel widget)
- `ksnip` (for the upload addon)
- Access to a SysReptor instance

---

### Installation

#### 1. Authorization Setup

The widget reads `server`, `token`, and `project_id` from the sysreptor CLI's default config file:

```
~/.sysreptor/config.yaml
```

If you use the sysreptor CLI (`reptor`), this file already exists. Verify with:

```bash
cat ~/.sysreptor/config.yaml
```

Expected content:
```yaml
server: http://reptor-local:8080
token: your_api_token_here
project_id: active-project-uuid
```

**If you do NOT use the sysreptor CLI**, uncomment the `server` and `token` lines in the widget config file (see section below).

API tokens can be generated at: **SysReptor → User → API Tokens**

---

#### 2. Widget Configuration File

Copy the example and edit:

```bash
mkdir -p ~/.config/reptor-widget
cp config.conf.example ~/.config/reptor-widget/config.conf
nano ~/.config/reptor-widget/config.conf
```

Available fields:

| Field | Description | Default |
|---|---|---|
| `language` | Dialog language: `pt-br` or `en-us` | `en-us` |
| `template_uuid` | UUID of the template used when creating new findings | — |
| `api_upload_field` | Name of the Markdown field in the project Design | `api_upload` |
| `server` | *(optional)* SysReptor URL — use if you don't have the sysreptor CLI | — |
| `token` | *(optional)* API Token — use if you don't have the sysreptor CLI | — |

**How to get `template_uuid`:**
1. Open SysReptor in your browser
2. Go to **Findings → Templates** and choose your template
3. The UUID is in the URL:
   ```
   http://reptor-local:8080/templates/7b0dd839-259a-4fb1-ad08-179ec2514a95/
                                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                      this is your template_uuid
   ```

**What is `api_upload_field`:**  
A **Markdown** type field in your SysReptor project Design. This is where ksnip-uploaded images accumulate. Create an internal field (not published in the final PDF) in your Design — for example `api_upload`. The name you set here must exactly match the field name in the Design.

---

#### 3. Build and Install

**Panel Widget:**
```bash
go build -o reptor_widget main.go
sudo cp reptor_widget /usr/local/bin/
```

**ksnip Addon:**
```bash
cd ksnip_addon_reptor
go build -o reptor-upload ksnip-reptor.go
sudo cp reptor-upload /usr/local/bin/
```

---

#### 4. Set Up the Panel Widget in XFCE

1. Right-click the panel → **Add New Items** → **Generic Monitor (GenMon)**
2. Right-click the added GenMon → **Properties**
3. Configure:
   - **Command:** `/usr/local/bin/reptor_widget`
   - **Period:** `30` (seconds between refreshes)
   - **Run in terminal:** unchecked
4. The panel will show the active project name in green. Click it to open the project selector.

---

#### 5. Configure ksnip

1. Open ksnip → **Options → Settings → Script**
2. In the **Post-capture script** field, enter:
   ```
   /usr/local/bin/reptor-upload %1
   ```
   `%1` is replaced by the path of the file saved by ksnip.

---

### Usage

**Switch project:** click the widget in the panel → select the project from the list.

**Send evidence:** capture a screenshot with ksnip → when saved, a dialog appears with the findings list for the active project → select an existing finding or create a new one → the image is uploaded and the Markdown field is updated.

---

### Author

Developed by **Luiz Le Fort** — aka LeFoFo  
Security researcher & pentester  
Built for personal use during red team engagements — shared with the community.
