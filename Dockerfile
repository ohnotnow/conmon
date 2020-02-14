FROM python:3.7-slim

LABEL maintainer="billybofh@gmail.com"

WORKDIR /conmon
COPY conmon.py requirements.txt /conmon/
RUN pip install -r requirements.txt
CMD ["python", "-u", "conmon.py"]
