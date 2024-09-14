FROM alpine:latest

RUN apk add --no-cache curl

RUN curl -sSf https://atlasgo.sh | sh
    
ENV DATABASE_URL=""

COPY src/AuthService.WebApi/atlas/migrations /migrations 


# Force re-hash, there will be a pipeline to verifiy the migrations integrity
# this was needed cuz the git was altering the files endlines
RUN exec atlas migrate hash

CMD exec atlas migrate apply --dir file:///migrations --url $DATABASE_URL