FROM python:3.8

WORKDIR /workspace
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .

ENTRYPOINT ["python", "image.py"]