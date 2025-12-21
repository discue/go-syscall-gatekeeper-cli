FROM golang

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY read.py /read.py
RUN apt-get update && apt-get install -y python3 && rm -rf /var/lib/apt/lists/*

CMD ["python3", "/read.py"]