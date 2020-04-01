FROM debian:bullseye

WORKDIR /usr/local/bin
ADD k8saadmsi k8saadmsi

CMD ["k8saadmsi"]
