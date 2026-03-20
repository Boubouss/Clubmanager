run: 
	@docker compose up

clean:
	@docker compose down -v --rmi all

re:
	@docker compose down --rmi all
	make run
