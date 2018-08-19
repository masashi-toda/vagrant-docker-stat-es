# https://github.com/elastic/elasticsearch-docker
# https://www.elastic.co/guide/en/elasticsearch/reference/current/docker.html
#FROM docker.elastic.co/elasticsearch/elasticsearch-oss:6.1.1
FROM docker.elastic.co/elasticsearch/elasticsearch:6.1.1

# Add your elasticsearch plugins setup here
RUN elasticsearch-plugin install analysis-icu
RUN elasticsearch-plugin install analysis-kuromoji
