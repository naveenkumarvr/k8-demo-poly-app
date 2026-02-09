import React from 'react';
import { CheckCircle, X } from 'lucide-react';

export default function CheckoutSuccessModal({ transaction, onClose }) {
    if (!transaction) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/50"
                onClick={onClose}
            />

            {/* Modal */}
            <div className="relative bg-white rounded-2xl shadow-2xl max-w-md w-full p-6 transform transition-all">
                <button
                    onClick={onClose}
                    className="absolute top-4 right-4 p-2 hover:bg-gray-100 rounded-full transition-colors"
                    aria-label="Close"
                >
                    <X className="w-5 h-5 text-gray-500" />
                </button>

                <div className="text-center">
                    <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4">
                        <CheckCircle className="w-10 h-10 text-green-600" />
                    </div>

                    <h2 className="text-2xl font-bold text-gray-900 mb-2">
                        {transaction.status === 'SUCCESS' ? 'Order Successful!' : 'Order Status'}
                    </h2>

                    <p className="text-gray-600 mb-6">
                        {transaction.message || 'Your order has been processed'}
                    </p>

                    <div className="bg-gray-50 rounded-lg p-4 mb-6 space-y-3">
                        <div className="flex justify-between items-center">
                            <span className="text-sm font-medium text-gray-600">Transaction ID</span>
                            <span className="text-sm font-mono bg-white px-3 py-1 rounded border text-gray-900">
                                {transaction.transactionId || transaction.transaction_id}
                            </span>
                        </div>

                        <div className="flex justify-between items-center">
                            <span className="text-sm font-medium text-gray-600">Status</span>
                            <span className={`text-sm font-semibold px-3 py-1 rounded ${transaction.status === 'SUCCESS'
                                    ? 'bg-green-100 text-green-700'
                                    : 'bg-yellow-100 text-yellow-700'
                                }`}>
                                {transaction.status}
                            </span>
                        </div>

                        {transaction.totalItems !== undefined && (
                            <div className="flex justify-between items-center">
                                <span className="text-sm font-medium text-gray-600">Total Items</span>
                                <span className="text-sm font-semibold text-gray-900">
                                    {transaction.totalItems || transaction.total_items}
                                </span>
                            </div>
                        )}
                    </div>

                    <button
                        onClick={onClose}
                        className="w-full bg-blue-600 text-white font-semibold py-3 px-6 rounded-lg hover:bg-blue-700 transition-colors"
                    >
                        Continue Shopping
                    </button>
                </div>
            </div>
        </div>
    );
}
