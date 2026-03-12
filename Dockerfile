# SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
# SPDX-License-Identifier: MIT

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY cheetah /usr/local/bin/cheetah

ENTRYPOINT ["/usr/local/bin/cheetah"]
CMD ["--help"]
