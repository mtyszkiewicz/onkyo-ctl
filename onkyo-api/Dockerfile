FROM python:3.11

WORKDIR /app

RUN pip install poetry

COPY poetry.lock pyproject.toml ./

RUN poetry install

COPY . .

CMD ["poetry", "run", "uvicorn", "onkyo_api.main:app", "--host", "0.0.0.0", "--port", "8000"]
