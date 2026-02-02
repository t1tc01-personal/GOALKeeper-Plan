from fastapi import APIRouter
from handler.hello_world import HelloWorldHandler


class HelloWorldRoute:
    router: APIRouter
    handler: HelloWorldHandler

    def __init__(self, handler: HelloWorldHandler):
        self.router = APIRouter()
        self.handler = handler

        self.router.add_api_route(
            path="/",
            endpoint=self.handler.hello_world,
            methods=["GET"],
            response_model=str,
            summary="Hello World",
            description="Hello World",
        )
