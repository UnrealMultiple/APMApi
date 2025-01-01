import io
import json
from glob import glob

import uvicorn
from starlette.responses import FileResponse

from fastapi import FastAPI, File, UploadFile, Form, HTTPException
from fastapi.responses import JSONResponse
from config import settings
import os
import shutil
import zipfile

app = FastAPI()
uploaded_plugins_path = 'uploaded_plugins'
packed_plugins_path = 'packed_plugins'
plugin_list: json
plugin_list = None


def internal_get_plugin_list():
    global plugin_list
    if plugin_list is None:
        with open("uploaded_plugins/Plugins.json", 'r', encoding='utf-8') as file:
            plugin_list = json.loads(file.read())
    return plugin_list


def packet_plugin(assembly_name: str):
    file_paths = glob(f"{uploaded_plugins_path}/Plugins/{assembly_name}.*")
    with zipfile.ZipFile(f"{packed_plugins_path}/{assembly_name}.zip", 'w') as zip_archive:
        for file_path in file_paths:
            with open(file_path, 'rb') as f:
                data = f.read()
                zip_archive.writestr(os.path.basename(file_path), data)
    print(f"{assembly_name + '.zip'}打包成功.")


@app.get("/supermarket/xml")
async def supermarket_xml():
    return {"备案号": "闽ICP备2024057933号-1", "好神秘": "为什么你会想去看这个API捏?"}


@app.post("/plugin/upload")
async def upload_plugin_zip(token: str = Form(...), file: UploadFile = File(...)):
    if token != settings.token:
        raise HTTPException(status_code=401, detail="没认证捏~")
    content = await file.read()
    print(f"收到插件包: {file.filename}")
    print(f"大小: {len(content)} bytes")

    if os.path.exists(uploaded_plugins_path):
        shutil.rmtree(uploaded_plugins_path)
    os.makedirs(uploaded_plugins_path)

    if os.path.exists(packed_plugins_path):
        shutil.rmtree(packed_plugins_path)
    os.makedirs(packed_plugins_path)

    with zipfile.ZipFile(io.BytesIO(content), 'r') as zip_ref:
        zip_ref.extractall(uploaded_plugins_path)

    with open(uploaded_plugins_path + '/Plugins.zip', 'wb') as file:
        file.write(content)

    global plugin_list
    with open("uploaded_plugins/Plugins.json", 'r', encoding='utf-8') as file:

        plugin_list = json.loads(file.read())

    for p in plugin_list:
        packet_plugin(p['AssemblyName'])
    print(f"已解压插件包到: {uploaded_plugins_path}")

    return JSONResponse(content={"message": "插件包更新成功~"})


@app.get("/plugin/get_plugin_list")
async def get_plugin_list():
    return JSONResponse(internal_get_plugin_list())


@app.get("/plugin/get_all_plugins")
async def get_all_plugins():
    return FileResponse(f'{uploaded_plugins_path}/Plugins.zip', filename="Plugins.zip")


@app.get("/plugin/get_plugin_zip")
async def get_plugin_zip(assembly_name):
    return FileResponse(f'{packed_plugins_path}/{assembly_name}.zip', filename=f'{assembly_name}.zip')


if __name__ == '__main__':
    uvicorn.run(app, host="0.0.0.0", port=11434)
