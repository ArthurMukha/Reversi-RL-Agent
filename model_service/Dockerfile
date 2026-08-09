FROM python:3.14-slim

WORKDIR /app

COPY model_service/requirements.txt .

RUN pip install -r requirements.txt

COPY model_service/ /app/model_service/

ENV LISTEN_ADDR=0.0.0.0:50051

EXPOSE 50051

CMD ["python", "-m", "model_service.server"]