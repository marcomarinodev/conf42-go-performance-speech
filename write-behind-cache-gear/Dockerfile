FROM redis:latest

RUN apt-get update && apt-get install -y make gcc

# Download and compile Redis modules
RUN git clone https://github.com/redislabs/redismod.git
RUN cd redismod && make
RUN cp modules/*.so /usr/local/lib/redis/

# (Optional) Configure Redis to load modules automatically
RUN echo "loadmodule /usr/local/lib/redis/redisgears.so" >> /etc/redis/redis.conf

EXPOSE 6379

CMD ["redis-server"]
