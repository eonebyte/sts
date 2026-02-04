export interface HandoverRequest {
  m_inout_ids?: number[];
  status: string;
  user_id: number;
  driver_by?: number;
  customer_id?: number;
  tnkb_id?: number;
  current_customer?: number;
  notes?: string;
}

export interface ApiResponse {
  success: boolean;
  message: string;
  data?: any;
}

export interface SuratJalanItem {
  m_inout_id: number;
  document_no: string;
  customer_name: string;
  status: string;
  tnkb_no: string; // Konsisten dengan JSON response
  tnkb_id: string; // Konsisten dengan JSON response
  customer_id: string; // Konsisten dengan JSON response
  movement_date?: string;
  driver_name?: string;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const apiService = {
  getToken() {
    return typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  },

  async getActiveShipments(driverId: number): Promise<SuratJalanItem[]> {
    const token = this.getToken();
    const response = await fetch(`${API_BASE_URL}/shipments/in-transit?driverId=${driverId}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Gagal mengambil data shipment`);
    }

    const result = await response.json();

    // Mapping disesuaikan dengan JSON response: 
    // { "m_inout_id": 1515551, "document_no": "860051", "tnkb_no": "AB 8856 AK", ... }
    return (result.data || []).map((item: any) => ({
      m_inout_id: item.m_inout_id,
      document_no: item.document_no,
      customer_id: item.customer_id,
      customer_name: item.customer_name,
      movement_date: item.movement_date,
      driver_name: item.driver_name,
      tnkb_id: item.tnkb_id,
      tnkb_no: item.tnkb_no, // Gunakan tnkb_no
    }));
  },

  async getOnCustomerShipments(customerId: number, driverId: number): Promise<SuratJalanItem[]> {
    const token = this.getToken();
    const response = await fetch(`${API_BASE_URL}/shipments/on-customer?customerId=${customerId}&driverId=${driverId}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Gagal mengambil data shipment`);
    }

    const result = await response.json();

    // Mapping disesuaikan dengan JSON response: 
    // { "m_inout_id": 1515551, "document_no": "860051", "tnkb_no": "AB 8856 AK", ... }
    return (result.data || []).map((item: any) => ({
      m_inout_id: item.m_inout_id,
      document_no: item.document_no,
      customer_id: item.customer_id,
      customer_name: item.customer_name,
      status: item.status,
      movement_date: item.movement_date,
      driver_name: item.driver_name,
      tnkb_id: item.tnkb_id,
      tnkb_no: item.tnkb_no, // Gunakan tnkb_no
    }));
  },

  async processHandover(req: HandoverRequest): Promise<ApiResponse> {
    const token = this.getToken();
    const response = await fetch(`${API_BASE_URL}/handover/process`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(req),
    });

    return response.json();
  },
};