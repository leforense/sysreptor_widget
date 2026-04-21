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
   - **Optional:** Check "use a sigle panel row"
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

### Demonstração e configuração com imagens

#### Configurando o Widget

1. (opcional) Edite seu Design template e inclua um novo campo (ele não precisa ser publicado ao gerar o PDF), como o exemplo abaixo: [Caso não queira incluir um novo campo, poderá usar um já existente. o Script não apaga o conteúdo, ele acrescenta no existente]
<img width="1717" height="746" alt="image" src="https://github.com/user-attachments/assets/7f5a0fd6-694d-4f02-b5d5-b6877f231d59" />


2. (necessario) defina um template de finding "em branco", ou pré-preenchido, conforme exemplo, o importante aqui é a UUiD

<img width="784" height="262" alt="image" src="https://github.com/user-attachments/assets/a675b2fa-eb5d-4cb6-a2bd-58b77674b6fd" />

3. Após compilar e copiar os arquivos binário, configure o Widget para Kali, pelos passos abaixo:
- Botão direito sobre o painel do Kali (xfce4-panel), Panel --> Panel Preferences
- Selecione (Itens)
- Selecione Generic Monitor e clique em Add e clique em Close

<img width="791" height="760" alt="image" src="https://github.com/user-attachments/assets/cda9a824-636f-43c8-82da-40181527e057" />

- Ainda no Panel Preferences, role até o último item, clique sobre Generic Monitor, clique em (3 traços) e adicione `/usr/local/bin/reptor_widget`, conforme exemplo:
<img width="736" height="766" alt="image" src="https://github.com/user-attachments/assets/fb2e6c2d-828d-46ff-a92e-eafb56a74f34" />
- Se tudo estiver correto, será exibido algo similar a isto:
<img width="255" height="70" alt="image" src="https://github.com/user-attachments/assets/86f58728-3993-467b-9601-0379f7c8b1ac" />
- (Opcional) mova o Generic Monitor para a posição desejada no seu painel.

4. Com o widget posicionado, agora só clicar em cima dele, para trocar de projeto conforme exemplo.
<img width="596" height="470" alt="image" src="https://github.com/user-attachments/assets/07978933-bd06-4c1d-8de8-90f737b3ab9c" />

Nota: Se a conexão estiver indisponível, ficará como "Reptor not connected"

#### Configurando o ksnip e seu funcionamento

NOTA: (não é obrigatório o widget para o plugin funcionar, no entanto, tenha atenção redobrada durante o upload de imagens)

1. Após compilar e copiar o binário para `/usr/local/bin/reptor-upload` abra seu KSNIP (Pressumindo que já está instalado)
2. Abra seu KSNIP, e configure:
- Options --> Settings
- No menu a esquerda, selecione (Script Uploader) e ao lado da direita coloque o local do binário, conforme exemplo:
<img width="820" height="673" alt="image" src="https://github.com/user-attachments/assets/fadb4614-6045-4c59-bdd9-021f8b966a14" />

3. (OPCIONAL), esta interação foi realizada para evitar (cliques inuteis do Ksnip)
- Clique em Actions --> Upload --> Defina um nome `Ex. Upload`, defina um short-cut, defina como `Global`, marque a caixa de selação `Take Capture` e marque a caixa de seleção `Upload Image`, demais itens utilize de acordo com suas preferências de trabalho.
<img width="829" height="685" alt="image" src="https://github.com/user-attachments/assets/3303a861-1bd9-441a-ae8c-9aa759e0e656" />

4. Agora, quando você pressionar `shift+print` sera capturado a região, e o ksnip vai chamar o `reptor-upload` com a seguinte interação:
<img width="614" height="415" alt="image" src="https://github.com/user-attachments/assets/5d2910ce-c4a9-4a20-b2af-d203ed3ad48b" />

5. Após selecionar o finding correspondente da evidência, via API será enviado para o campo pré-definido em cofig.cong (demostrado no ponto 1 de configuração do Widget), conforme exemplo:
<img width="1216" height="734" alt="image" src="https://github.com/user-attachments/assets/0bf4b504-7a55-4211-a329-bfb506807033" />

Notas:
- Este programa, memoriza o ultimo finding utilizado (para envio em massa)
- Caso voce mude de projeto, ele limpa o ultimo utilizado
- Caso exista conteúdo no campo definido, não será sobrescrito, será **ACRESCENTADO** conforme este exemplo:
<img width="1211" height="901" alt="image" src="https://github.com/user-attachments/assets/4f06cca8-fc45-4a1f-a193-631b2da4aa9a" />
- Caso seja um novo FINDING, poderá usar o template (em branco, definido no passo 2 do Widget)


--- Bugs conhecidos!

Se no arquivo `~/.sysreptor/config.yaml`, este linha "project_id: bfd1ed3a-ccfd-4eea-b3ea-36ae1f4d52fa" estiver em branco, ou com um projeto inexistente, será exibido como "Reptor not connected", para corrigir, basta colocar um projeto válido, editando o arquivo ou usando `reptor project` 



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

### Demonstration and configuration with images

#### Configuring the Widget

1. (optional) Edit your Design template and add a new field (it does not need to be published when generating the PDF), as shown below: [If you don't want to add a new field, you can use an existing one. The script does not overwrite the content — it **appends** to it]
<img width="1717" height="746" alt="image" src="https://github.com/user-attachments/assets/7f5a0fd6-694d-4f02-b5d5-b6877f231d59" />

2. (required) Define a finding template — blank or pre-filled — as shown; what matters here is the UUID
<img width="784" height="262" alt="image" src="https://github.com/user-attachments/assets/a675b2fa-eb5d-4cb6-a2bd-58b77674b6fd" />

3. After compiling and copying the binary files, configure the Widget for Kali as follows:
- Right-click the Kali panel (xfce4-panel), Panel → Panel Preferences
- Select (Items)
- Select Generic Monitor, click Add, then click Close

<img width="791" height="760" alt="image" src="https://github.com/user-attachments/assets/cda9a824-636f-43c8-82da-40181527e057" />

- Still in Panel Preferences, scroll to the last item, click Generic Monitor, click (3 lines) and add `/usr/local/bin/reptor_widget`, as shown:
<img width="736" height="766" alt="image" src="https://github.com/user-attachments/assets/fb2e6c2d-828d-46ff-a92e-eafb56a74f34" />
- If everything is correct, something similar to this will be displayed:
<img width="255" height="70" alt="image" src="https://github.com/user-attachments/assets/86f58728-3993-467b-9601-0379f7c8b1ac" />
- (Optional) Move the Generic Monitor to the desired position on your panel.

4. With the widget in place, just click on it to switch projects as shown.
<img width="596" height="470" alt="image" src="https://github.com/user-attachments/assets/07978933-bd06-4c1d-8de8-90f737b3ab9c" />

Note: If the connection is unavailable, it will display "Reptor not connected"

#### Configuring ksnip and its operation

NOTE: (the widget is not required for the plugin to work, however, pay close attention to which project is active when uploading images)

1. After compiling and copying the binary to `/usr/local/bin/reptor-upload`, open KSNIP (assuming it is already installed)
2. Open KSNIP and configure:
- Options → Settings
- In the left menu, select (Script Uploader) and on the right side enter the binary path, as shown:
<img width="820" height="673" alt="image" src="https://github.com/user-attachments/assets/fadb4614-6045-4c59-bdd9-021f8b966a14" />

3. (OPTIONAL) This interaction was created to avoid unnecessary ksnip clicks:
- Click Actions → Upload → Set a name (e.g. `Upload`), define a shortcut, set as `Global`, check `Take Capture` and check `Upload Image`; use the remaining options according to your work preferences.
<img width="829" height="685" alt="image" src="https://github.com/user-attachments/assets/3303a861-1bd9-441a-ae8c-9aa759e0e656" />

4. Now, when you press `shift+print`, the region will be captured and ksnip will call `reptor-upload` with the following interaction:
<img width="614" height="415" alt="image" src="https://github.com/user-attachments/assets/5d2910ce-c4a9-4a20-b2af-d203ed3ad48b" />

5. After selecting the corresponding finding for the evidence, it will be sent via API to the pre-defined field in `config.conf` (shown in Widget configuration step 1), as shown:
<img width="1216" height="734" alt="image" src="https://github.com/user-attachments/assets/0bf4b504-7a55-4211-a329-bfb506807033" />

Notes:
- This program remembers the last finding used (for bulk uploads)
- If you switch projects, it clears the last used finding
- If the defined field already has content, it will NOT be overwritten — it will be **APPENDED**, as in this example:
<img width="1211" height="901" alt="image" src="https://github.com/user-attachments/assets/4f06cca8-fc45-4a1f-a193-631b2da4aa9a" />
- If it is a new FINDING, you can use the blank template defined in Widget step 2

---

### Known Bugs

If in `~/.sysreptor/config.yaml` the line `project_id:` is blank or contains a non-existent project UUID, the widget will display "Reptor not connected". To fix this, set a valid project by editing the file directly or running `reptor project`.

---

### Author

Developed by **Luiz Le Fort** — aka LeFoFo / leforense  
Security researcher & pentester  
Built for personal use during red team engagements — shared with the community.
