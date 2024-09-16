# API em Go (ClassAPI) - Go 1.23.0

Este projeto é uma API simples para cadastro dos alunos da univercidade unifg desenvolvida com GoLang que utiliza MongoDB como banco de dados. As instruções a seguir orientam sobre como configurar e rodar o projeto.

## Pré-requisitos

- **Go**: Certifique-se de que o Go 1.23.0 ou superior está instalado na sua máquina. Se não tiver, faça o download [aqui](https://go.dev/dl/).
- **MongoDB**: Ter uma instância do MongoDB rodando. Você pode usar o [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) ou uma instância local do MongoDB.

## Configuração do Projeto

1. Clone este repositório:

    ```bash
    git clone https://github.com/ADS-Unifg/ClassApi
    cd ClassApi
    ```

2. Remova o sufixo `-example` do arquivo `.env-example`:

    ```bash
    mv .env-example .env
    ```

3. Edite o arquivo `.env` para adicionar a URL de conexão com o MongoDB. O conteúdo do arquivo deve se parecer com:

    ```bash
    MONGO_URI=mongodb+srv://<usuario>:<senha>@<cluster-url>/<nome-do-banco>?retryWrites=true&w=majority
    ```

4. Execute o comando para resolver as dependências do projeto:

    ```bash
    go mod tidy
    ```

## Rodando a Aplicação

1. Para iniciar a aplicação, execute o comando:

    ```bash
    go run main.go
    ```

2. A aplicação será iniciada e estará disponível no endereço [http://localhost:8080](http://localhost:80).

## Contato

Se precisar de ajuda, entre em contato
