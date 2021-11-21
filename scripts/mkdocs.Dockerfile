FROM python:3
RUN pip install mkdocs && pip install mkdocs-ivory

WORKDIR /build

ENTRYPOINT ["mkdocs"]
CMD ["build"]
