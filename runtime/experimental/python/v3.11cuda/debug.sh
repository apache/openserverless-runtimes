#!/bin/bash
export OW_LOG_INIT_ERROR=1 OW_WAIT_FOR_ACK=1 OW_COMPILER=/bin/compile OW_ACTIVATE_PROXY_SERVER=1
proxy -debug
