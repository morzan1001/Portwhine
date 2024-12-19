import docker

class DockerManager:
    def __init__(self):
        self.client = docker.from_env()

    def start_container(self, container_name, image_name, command=None, **kwargs):
        try:
            container = self.client.containers.run(
                image_name, 
                name=container_name, 
                command=command, 
                detach=True, 
                **kwargs
            )
            print(f"Started container {container_name} using image {image_name} with command '{command}'.")
            return container
        except docker.errors.APIError as e:
            print(f"Error starting container {container_name}: {e}")

    def stop_container(self, container_name):
        try:
            container = self.client.containers.get(container_name)
            if container.status == "running":
                container.stop()
                container.remove()
                print(f"Stopped and removed {container_name} container.")
            else:
                print(f"{container_name} container is not running.")
        except docker.errors.NotFound:
            print(f"{container_name} container not found.")
        except docker.errors.APIError as e:
            print(f"Error stopping container {container_name}: {e}")

# Beispiel für die Verwendung des Controllers
if __name__ == "__main__":
    manager = DockerManager()
    manager.start_container("my_container", "nmap", command="nmap -sP 192.168.1.0/24", network="bridge")
    manager.stop_container("my_container")