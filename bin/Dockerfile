FROM alpine:3.9

RUN apk add bind-tools
RUN  apk update && \
     apk add libc6-compat && \
     apk add ca-certificates && \
     apk add logrotate
     
ARG commit=commit
ENV commit=$commit


COPY ./bootnode /
COPY ./keylist.json /
COPY ./keylist_256.json /
COPY ./sample-config.conf /


RUN chmod +x /bootnode

COPY ./run_bootnode.sh /
COPY ./run_incognito.sh /
COPY ./run_incognito_dev.sh /
COPY ./run_fullnode.sh /


RUN chmod +x /run_bootnode.sh
RUN chmod +x /run_incognito_dev.sh
RUN chmod +x /run_fullnode.sh

RUN mkdir /utility

#COPY ./txs-shard0-noprivacy-5000.json /utility/
#COPY ./txs-shard0-noprivacy-9000.json /utility/
#COPY ./txs-shard0-noprivacy-10000.2.json /utility/
#COPY ./txs-shard0-noprivacy-10000.3.json /utility/
#COPY ./txs-shard0-noprivacy-10000.4.json /utility/
#COPY ./txs-shard0-noprivacy-10000.5.json /utility/
#COPY ./txs-shard0-privacy-5000.json /utility/
#COPY ./txs-shard0-privacy-3000.1.json /utility/
#COPY ./txs-shard0-privacy-3000.2.json /utility/
#COPY ./txs-shard0-privacy-3000.3.json /utility/

#COPY ./txs-shard1-noprivacy-5000.json /utility/
#COPY ./txs-shard1-noprivacy-9000.json /utility/
#COPY ./txs-shard1-noprivacy-10000.2.json /utility/
#COPY ./txs-shard1-noprivacy-10000.3.json /utility/
#COPY ./txs-shard1-noprivacy-10000.4.json /utility/
#COPY ./txs-shard1-noprivacy-10000.5.json /utility/
#COPY ./txs-shard1-cstoken-5000.json /utility/
#COPY ./txs-shard1-cstokenprivacy-5000.json /utility/
#COPY ./txs-shard1-privacy-5000.json /utility/
#COPY ./txs-shard1-privacy-3000.1.json /utility/
#COPY ./txs-shard1-privacy-3000.2.json /utility/
#COPY ./txs-shard1-privacy-3000.3.json /utility/
CMD ["/bin/sh","run_incognito.sh"]

ARG commit=commit
ENV commit=$commit

COPY ./incognito /
RUN chmod +x /incognito
