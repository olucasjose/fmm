# FMM (Fastest Mint Mirrors)

🌍 *Leia em outros idiomas: [English](README.md) | [Português](README.pt-BR.md)*

FMM é uma ferramenta rápida e interativa para rankear e selecionar mirrors (espelhos) do Linux Mint. Ela foi desenvolvida para testar rapidamente os mirrors disponíveis e ajudar você a atualizar a configuração de mirrors do seu sistema para otimizar as velocidades de download.

## Escopo e Filosofia

**O FMM foca exclusivamente no teste e na atualização de mirrors.**

Outros recursos encontrados na ferramenta original `mintsources` — como o gerenciamento de chaves GPG, gerenciamento de PPAs e outras configurações de repositório — **não** fazem parte do escopo deste projeto. Esses recursos não foram implementados e não há planos para implementá-los no futuro. O FMM foi construído para fazer apenas uma coisa e fazê-la extremamente bem: encontrar e configurar os mirrors mais rápidos.

## Origem e Créditos

O FMM usa o projeto original [mintsources](https://github.com/linuxmint/mintsources) como sua base direta. Este projeto é um port em Go e um trabalho derivado da lógica de testes de mirrors do `mintsources`, que é desenvolvido e mantido pela equipe do Linux Mint. O FMM baseia-se nessa estrutura robusta para fornecer uma experiência de linha de comando (CLI) mais rápida.

O FMM é licenciado sob a **GNU General Public License v3.0 (GPLv3)**. O software original `mintsources` é distribuído sob a licença GNU GPL. Consulte o arquivo [LICENSE](LICENSE) para obter mais detalhes.

## Instalação

Você pode instalar o FMM diretamente executando o script de instalação fornecido. O script compila o binário em Go, move-o para o caminho do seu sistema e configura o autocompletar no bash.

```bash
# Clone o repositório
git clone https://github.com/olucasjose/fmm.git
cd fmm

# Execute o script de instalação (requer privilégios sudo)
./install.sh
```

## Como Usar

O FMM exige privilégios `sudo` para ações que modificam as configurações do sistema (como `run`), mas os comandos informativos podem ser executados por um usuário comum.

- **Execute o fluxo principal de teste e atualização de mirrors:**
  ```bash
  sudo fmm run
  ```

- **Liste os mirrors disponíveis:**
  ```bash
  fmm list
  ```

- **Mostre o ranking atual de mirrors (apenas testa sem aplicar no sistema):**
  ```bash
  fmm ranking
  ```

- **Exiba a ajuda e os comandos disponíveis:**
  ```bash
  fmm help
  ```
