run: 
	@docker compose up

clean:
	@docker compose down -v --rmi all

re:
	make clean
	make run
