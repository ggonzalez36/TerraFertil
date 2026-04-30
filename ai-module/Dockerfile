FROM python:3.11-slim



ENV PYTHONDONTWRITEBYTECODE=1

ENV PYTHONUNBUFFERED=1



WORKDIR /app



COPY requirements.txt ./

RUN pip install --no-cache-dir -r requirements.txt



COPY app ./app

COPY scoring ./scoring



ENV BACKEND_API_URL=http://backend-go:8080

ENV AI_DATABASE_PATH=/data/ai_module.db



RUN useradd -m app && mkdir -p /data && chown -R app:app /app /data

USER app



EXPOSE 8001



CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8001"]