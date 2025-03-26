defmodule TCPClient do
  def connect(host, port) do
    case :gen_tcp.connect(host, port, [:binary, active: false, packet: :line]) do
      {:ok, socket} ->
        spawn(fn -> listen_loop(socket) end)
        command_loop(socket)
      {:error, reason} ->
        IO.puts("Connection failed: #{inspect(reason)}")
        {:error, reason}
    end
  end

  defp command_loop(socket) do
    case IO.gets("> ") do
      :eof ->
        :gen_tcp.close(socket)
      data ->
        :gen_tcp.send(socket, data)
        command_loop(socket)
    end
  end

  defp listen_loop(socket) do
    case :gen_tcp.recv(socket, 0) do # recv since it's passive connect
      {:ok, msg} ->
        IO.write(msg)
        listen_loop(socket)
      {:error, :closed} ->
        IO.puts("Socket closed")
      {:error, reason} ->
        IO.puts("Error receiving data: #{inspect(reason)}")
    end
  end
end

with [host, port | _] <- System.argv(),
      {port, ""} <- Integer.parse(port) do
  TCPClient.connect(String.to_atom(host), port)
else
  _ -> TCPClient.connect(~c"localhost", 6666)
end
