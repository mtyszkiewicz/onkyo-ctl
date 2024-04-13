FROM fnndsc/python-poetry:1.7.1

WORKDIR /app

COPY poetry.lock pyproject.toml ./

RUN poetry install

COPY onkyo_api ./

CMD ["poetry", "run", "uvicorn", "onkyo_api.main:app"]