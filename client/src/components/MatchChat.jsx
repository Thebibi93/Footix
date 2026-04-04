import { useEffect, useState } from "react";
import { useSimulatedApi } from "../hooks/useSimulatedApi";
import { useAuth } from "../contexts/AuthContext";

export default function MatchChat({ matchId, matchTitle }) {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const { getChatMessages, sendChatMessage } = useSimulatedApi();
  const { user } = useAuth();

  useEffect(() => {
    if (matchId) {
      getChatMessages(matchId).then(setMessages);
    }
  }, [matchId]);

  const handleSend = async () => {
    if (!newMessage.trim()) return;
    if (!user) return;
    const sent = await sendChatMessage(matchId, newMessage, user);
    setMessages((prev) => [sent, ...prev]);
    setNewMessage("");
  };

  return (
    <div className="mt-8 border-t pt-6">
      <h3 className="text-xl font-semibold mb-3">
        💬 Chat du match {matchTitle}
      </h3>
      <div className="bg-gray-50 rounded-lg shadow-inner p-4">
        <div className="h-64 overflow-y-auto space-y-2 mb-3 flex flex-col-reverse">
          {messages.length === 0 && (
            <p className="text-gray-400 text-center">
              Aucun message pour l'instant.
            </p>
          )}
          {messages.map((msg) => (
            <div key={msg.id} className="border-b pb-1 text-sm">
              <span className="font-semibold text-blue-600">{msg.user}</span>
              <span className="text-xs text-gray-400 ml-2">
                {msg.timestamp}
              </span>
              <p className="text-gray-800">{msg.message}</p>
            </div>
          ))}
        </div>
        {user ? (
          <div className="flex gap-2">
            <input
              type="text"
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
              className="flex-1 border rounded px-3 py-2 text-sm"
              placeholder="Écrivez votre message..."
              onKeyPress={(e) => e.key === "Enter" && handleSend()}
            />
            <button
              onClick={handleSend}
              className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 text-sm"
            >
              Envoyer
            </button>
          </div>
        ) : (
          <p className="text-center text-gray-500 text-sm">
            Connectez-vous pour participer au chat.
          </p>
        )}
      </div>
    </div>
  );
}
