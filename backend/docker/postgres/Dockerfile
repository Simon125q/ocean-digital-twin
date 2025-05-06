FROM postgis/postgis:15-3.5

LABEL maintainer="Custom PostgreSQL with PostGIS and pg_cron extensions"

# Install pg_cron
RUN apt-get update \
    && apt-get install -y curl \
    && apt-get -y install postgresql-15-cron

# Configure pg_cron
# RUN echo "shared_preload_libraries='pg_cron'" >> /usr/share/postgresql/postgresql.conf.sample

