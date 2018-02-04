#!/usr/bin/env bash

set -eou pipefail

service_name_camel=${1}
service_name_lower=$(echo ${1} | tr '[:upper:]' '[:lower:]')
service_name_upper=$(echo ${1} | tr '[:lower:]' '[:upper:]')

TEMPLATE_CAMEL='ServiceName'
TEMPLATE_LOWER='servicename'
TEMPLATE_UPPER='SERVICENAME'

working_dir=$(pwd)
bootstrap_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cp -r ${bootstrap_dir}/template/* ${working_dir}
cp -r ${bootstrap_dir}/template/.* ${working_dir}

# move directories and files
for f_tmpl in $(find ${working_dir} -type d -name 'servicename*'); do
    f=$(echo ${f_tmpl} | sed -e "s|${TEMPLATE_LOWER}|${service_name_lower}|g")
    mv ${f_tmpl} ${f}
done
for f_tmpl in $(find ${working_dir} -type f -name 'servicename*'); do
    f=$(echo ${f_tmpl} | sed -e "s|${TEMPLATE_LOWER}|${service_name_lower}|g")
    mv ${f_tmpl} ${f}
done

grep -lR "${TEMPLATE_LOWER}" . | grep -v '\.git' | xargs sed -i -e "s|${TEMPLATE_LOWER}|${service_name_lower}|g"
grep -lR "${TEMPLATE_UPPER}" . | grep -v '\.git' | xargs sed -i -e "s|${TEMPLATE_UPPER}|${service_name_upper}|g"
grep -lR "${TEMPLATE_CAMEL}" . | grep -v '\.git' | xargs sed -i -e "s|${TEMPLATE_CAMEL}|${service_name_camel}|g"
