#!/usr/bin/env python
"""
# Author: Nikhil Singh
# Email: nikhil.eltrx@gmail.com
# Purpose: Pull hosts data from Vectra brain using API and sve the output json file defined in conf file..
# Usage: 
##   - install Python 3
##   - configure get_hosts_conf.json
##   - 
#Compatiblity_tested: Python3, VEctra Brain: 7.1, API version : 2.2 :
"""
import requests
from requests.packages.urllib3.exceptions import InsecureRequestWarning
import json
import os
import logging

script_dir = os.path.realpath(os.path.dirname(__file__))
conf_file_name = 'conf_get_hosts.json'
logging.basicConfig(level=logging.INFO , filename=f'{script_dir}/get_hosts_log.log', filemode='w', format='%(name)s - %(levelname)s - %(message)s')


def get_conf():
    conf_file = f'{script_dir}/{conf_file_name}'
    f = open(conf_file)
    global conf
    conf = json.load(f)
    f.close()

def get_hosts():
    api_version =  conf.get('api_version', 'v2.2')
    vec_he = conf.get('vec_he', 'localhost')
    vec_base_url = f'https://{vec_he}/api/{api_version}'
    vec_host_url = vec_base_url + '/hosts'
    vec_auth_token = conf.get('vec_api_token')
    headers = {'Content-Type': 'application/json', 'Authorization': f'Token {vec_auth_token}'}
    payload = {'active_traffic': True, 'page_size' : conf.get('max_page_size',5000) , 'page': 0}
    
    ip_data = {}
    send_query = 'yes'
    requests.packages.urllib3.disable_warnings(InsecureRequestWarning)
    while send_query is not None:
        payload['page'] += 1 
        if conf.get('max_page_number', 500) <  payload['page']:
            logging.info('Stopping at Maximum page count of intetration: {}'.format(conf.get('max_page_number', 500)))
            break

        response = requests.get(url=vec_host_url, params=payload, verify=False, headers=headers)
        result = response.json()
        send_query = result.get('next')
        for h in result.get('results'):
            ip_data[h.get('last_source')] = { 'vec_host_id': h.get('id') , 'name': h.get('name') }
        
    logging.info('get_hosts Done! Total ip count : {}'.format(len(ip_data)))
    print('get_hosts Done! Total ip count : {}'.format(len(ip_data)))
    o = open(conf.get('output_file', 'output.json'), 'w')
    json.dump(ip_data, o)
    o.close()



if __name__ == '__main__':
    get_conf()
    get_hosts()
