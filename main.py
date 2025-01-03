import io
import json
import traceback
from glob import glob
import time

import aiofiles
import uvicorn
from starlette.responses import StreamingResponse

from fastapi import FastAPI, File, UploadFile, Form, HTTPException
from fastapi.responses import JSONResponse
from config import settings
import os
import shutil
import zipfile
from starlette.requests import Request

app = FastAPI()


@app.middleware("http")
async def dispatch(request: Request, call_next):
    start_time = time.time()
    response = await call_next(request)
    process_time = time.time() - start_time
    response.headers["X-Process-Time"] = str(process_time*1000)+"ms"
    return response

uploaded_plugins_path = 'uploaded_plugins'
packed_plugins_path = 'packed_plugins'
plugin_list: json
plugin_list = None
is_uploading = False

def internal_get_plugin_list():
    global plugin_list
    if plugin_list is None:
        with open("uploaded_plugins/Plugins.json", 'r', encoding='utf-8') as file:
            plugin_list = json.loads(file.read())
    return plugin_list


def packet_plugin(assembly_name: str):
    file_paths = glob(f"{uploaded_plugins_path}/Plugins/{assembly_name}.*")
    with zipfile.ZipFile(f"{packed_plugins_path}/{assembly_name}.zip", 'w', compresslevel=9) as zip_archive:
        for file_path in file_paths:
            with open(file_path, 'rb') as f:
                data = f.read()
                zip_archive.writestr(os.path.basename(file_path), data)
    # print(f"{assembly_name + '.zip'}打包成功.")


@app.get("/supermarket/xml")
async def supermarket_xml():
    return {"超(市)级感谢": "此下载镜像由迅猛龙赞助~",
            "备案号": "闽ICP备2024057933号-1", }


@app.post("/plugin/upload")
async def upload_plugin_zip(token: str = Form(...), file: UploadFile = File(...)):
    global is_uploading
    global plugin_list

    if token != settings.token:
        raise HTTPException(status_code=401, detail="没认证捏~")

    is_uploading = True
    try:
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


        with open("uploaded_plugins/Plugins.json", 'r', encoding='utf-8') as file:
            plugin_list = json.loads(file.read())

        for p in plugin_list:
            packet_plugin(p['AssemblyName'])
        print(f"插件包更新成功~")
    except Exception:
        print(f"插件包更新失败: {traceback.format_exc()}")
    finally:
        is_uploading = False

    return JSONResponse(content={"message": "插件包更新成功~"})


@app.get("/plugin/get_plugin_list")
async def get_plugin_list():
    return JSONResponse(internal_get_plugin_list())


@app.get("/plugin/get_plugin_manifest/")
async def get_plugin_list(assembly_name):
    _plugin_list = internal_get_plugin_list()
    for i in _plugin_list:
        if i['AssemblyName'] == assembly_name:
            return JSONResponse(i)

    return HTTPException(status_code=404, detail="没有找到这个插件捏~")


@app.get("/plugin/get_all_plugins")
async def get_all_plugins():
    file_path = f'{uploaded_plugins_path}/Plugins.zip'
    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail="插件包当前未上传,无法下载~")

    if is_uploading:
        raise HTTPException(403, "API正在同步插件中,请稍后重试~")

    file_size = os.path.getsize(file_path)
    async def file_stream():
        async with aiofiles.open(file_path, 'rb') as file:
              yield await file.read()

    return StreamingResponse(file_stream(), media_type="application/zip",
                             headers={"Content-Disposition": "attachment; filename=Plugins.zip","Content-Length": str(file_size)})




@app.get("/plugin/get_plugin_zip")
async def get_plugin_zip(assembly_name):
    file_path = f'{packed_plugins_path}/{assembly_name}.zip'

    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail="没有找到指定插件~")

    if is_uploading:
        raise HTTPException(403, "API正在同步插件中,请稍后重试~")


    file_size = os.path.getsize(file_path)

    async def file_stream():
        async with aiofiles.open(file_path, 'rb') as file:

                    yield await file.read()


    return StreamingResponse(file_stream(), media_type="application/zip",
                             headers={"Content-Disposition": f"attachment; filename={assembly_name}.zip",
                                      "Content-Length": str(file_size)})


if __name__ == '__main__':
    uvicorn.run(app, host="0.0.0.0", port=11434)

